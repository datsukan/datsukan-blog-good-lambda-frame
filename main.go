package frame

import (
	"encoding/json"
	"flag"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Response は正常系のレスポンスを定義した構造体
type Response struct {
	GoodCount int `json:"goodCount"`
}

// Response は異常系のレスポンスを定義した構造体
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

var headers = map[string]string{
	"Access-Control-Allow-Origin":  "*",
	"Access-Control-Allow-Methods": "GET",
	"Access-Control-Allow-Headers": "Content-Type",
}

var useCaseCore func(string) (int, error)

func Exec(ucc func(string) (int, error)) {
	useCaseCore = ucc

	t := flag.Bool("local", false, "ローカル実行か否か")
	ID := flag.String("id", "", "ローカル実行用の記事ID")
	flag.Parse()

	isLocal, err := isLocal(t, ID)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if isLocal {
		fmt.Println("local")
		localController(ID)
		return
	}

	fmt.Println("production")
	lambda.Start(controller)
}

// isLocal はローカル環境の実行であるかを判定する
func isLocal(t *bool, ID *string) (bool, error) {
	if !*t {
		return false, nil
	}

	if *ID == "" {
		fmt.Println("no exec")
		return false, fmt.Errorf("ローカル実行だがID指定が無いので処理不可能")
	}

	return true, nil
}

// localController はローカル環境での実行処理を行う
func localController(ID *string) {
	js, err := useCase(*ID)
	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Println(js)
}

// controller はAPI Gateway / AWS Lambda 上での実行処理を行う
func controller(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	articleID := request.PathParameters["article_id"]
	if articleID == "" {
		err := fmt.Errorf("article_id is empty")
		return responseBadRequestError(err)
	}

	js, err := useCase(articleID)
	if err != nil {
		return responseInternalServerError(err)
	}

	return responseSuccess(js)
}

// useCase はアプリケーションのIFに依存しないメインの処理を行う
func useCase(articleID string) (string, error) {
	a, err := useCaseCore(articleID)
	if err != nil {
		return "", nil
	}

	r := Response{
		GoodCount: a,
	}
	jb, err := json.Marshal(r)
	if err != nil {
		return "", err
	}

	return string(jb), nil
}

// responseBadRequestError はリクエスト不正のレスポンスを生成する
func responseBadRequestError(rerr error) (events.APIGatewayProxyResponse, error) {
	b := ErrorResponse{
		Error:   "bad request",
		Message: rerr.Error(),
	}
	jb, err := json.Marshal(b)
	if err != nil {
		r := events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    headers,
			Body:       err.Error(),
		}
		return r, nil
	}
	body := string(jb)

	r := events.APIGatewayProxyResponse{
		StatusCode: 400,
		Headers:    headers,
		Body:       body,
	}
	return r, nil
}

// responseInternalServerError はシステムエラーのレスポンスを生成する
func responseInternalServerError(rerr error) (events.APIGatewayProxyResponse, error) {
	b := ErrorResponse{
		Error:   "internal server error",
		Message: rerr.Error(),
	}
	jb, err := json.Marshal(b)
	if err != nil {
		r := events.APIGatewayProxyResponse{
			StatusCode: 500,
			Headers:    headers,
			Body:       err.Error(),
		}
		return r, nil
	}
	body := string(jb)

	r := events.APIGatewayProxyResponse{
		StatusCode: 500,
		Headers:    headers,
		Body:       body,
	}
	return r, nil
}

// responseSuccess は処理成功時のレスポンスを生成する
func responseSuccess(body string) (events.APIGatewayProxyResponse, error) {
	r := events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    headers,
		Body:       body,
	}
	return r, nil
}
