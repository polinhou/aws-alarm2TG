package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	json "github.com/goccy/go-json"
)

var bot *tgbotapi.BotAPI
var err error

var chatID string = os.Getenv("CHAT_ID")
var Token string = os.Getenv("TOKEN")

func getImage(trigger events.CloudWatchAlarmTrigger) ([]byte, error) {
	mySession := session.Must(session.NewSession())
	client := cloudwatch.New(mySession)

	metricWidget := fmt.Sprintf(`{"metrics": [[ "%s", "%s", "%s", "%s", {"stat": "%s" }]], "period": %d, "width": 800,"height": 600, "start": "-PT1H","end": "+PT3M","timezone": "+0800"}`,
		trigger.Namespace, trigger.MetricName, trigger.Dimensions[0].Name, trigger.Dimensions[0].Value, strings.ToUpper(trigger.Statistic[:1])+strings.ToLower(trigger.Statistic[1:]), trigger.Period)

	fmt.Println(metricWidget)

	req := &cloudwatch.GetMetricWidgetImageInput{
		MetricWidget: aws.String(metricWidget),
		OutputFormat: aws.String("png"),
	}

	resp, err := client.GetMetricWidgetImage(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return resp.MetricWidgetImage, nil
}

func handler(ctx context.Context, snsEvent events.SNSEvent) {

	for _, record := range snsEvent.Records {
		snsRecord := record.SNS

		var alarmMessage events.CloudWatchAlarmSNSPayload
		json.Unmarshal([]byte(snsRecord.Message), &alarmMessage)

		var alarmTrigger events.CloudWatchAlarmTrigger
		tiggerJson, _ := json.Marshal(alarmMessage.Trigger)
		json.Unmarshal(tiggerJson, &alarmTrigger)

		bot, err = tgbotapi.NewBotAPI(Token)
		if err != nil {
			log.Fatal(err)
		}
		bot.Debug = false

		message := fmt.Sprintf("<b>%s</b>\n%s", alarmMessage.AlarmName, alarmMessage.NewStateReason)

		fmt.Println(message)

		NewMsg := tgbotapi.NewMessageToChannel(chatID, message)
		NewMsg.ParseMode = tgbotapi.ModeHTML
		_, err := bot.Send(NewMsg)
		if err != nil {
			log.Fatal(err)
		}

		images, _ := getImage(alarmTrigger)

		photoFileBytes := tgbotapi.FileBytes{
			Name:  "picture",
			Bytes: images,
		}

		bot.Send(tgbotapi.NewPhotoToChannel(chatID, photoFileBytes))

	}
}

func main() {
	lambda.Start(handler)
}
