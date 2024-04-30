package main

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator"
	_ "github.com/lib/pq"
)

type BugOptions struct {
	Severity        string `json:"severity" validate:"required"`
	VersionAffected string `json:"version_affected" validate:"required"`
}

type FeatureRequestOptions struct {
	SponsorID int `json:"sponsor_id" validate:"required"`
}

type Issue struct {
	ID         int    `json:"id" validate:"required"`
	ReporterID int    `json:"reporter_id" validate:"required"`
	Priority   int    `json:"priority" validate:"required"`
	Status     string `json:"status" validate:"required"`
	*BugOptions
	*FeatureRequestOptions
}

func isBugIssue(issue Issue) bool {
	return issue.BugOptions != nil
}

func isFeatureRequestIssue(issue Issue) bool {
	return issue.FeatureRequestOptions != nil
}

func main() {
	// データベースの接続情報
	host := "localhost"
	port := 5432
	user := "postgres"
	password := "password"
	dbname := "test_database"

	// データベースに接続
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("データベースに接続できませんでした:", err)
		return
	}
	defer db.Close()

	// Issueテーブルを作成
	// 既に存在すれば作り直す
	// カラム: properties(jsonb)
	if _, err := db.Exec("DROP TABLE IF EXISTS issue"); err != nil {
		fmt.Println("テーブルを削除できませんでした:", err)
		return
	}
	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS issue (properties jsonb)"); err != nil {
		fmt.Println("テーブルを作成できませんでした:", err)
		return
	}

	// 実際のアプリケーションではHTTP requestでJSONデータが届く
	jsonData := `{
        "id": 1,
        "reporter_id": 11,
        "priority": 1,
        "status": "new",
        "severity": "low",
        "version_affected": "1.0.0"
    }`

	// 構造体への詰め替え
	issue := Issue{}
	if err := json.Unmarshal([]byte(jsonData), &issue); err != nil {
		fmt.Println("JSONデータを構造体に変換できませんでした")
		return
	}

	// バリデーション
	// 今回は必須チェックのみ
	validate := validator.New()
	if err := validate.Struct(issue); err != nil {
		fmt.Printf("バリデーションエラーです。 %#v\n", issue)
		return
	}

	// BugIssueかつFeatureRequestIssueはあり得ないという仕様
	if isBugIssue(issue) == isFeatureRequestIssue(issue) {
		fmt.Printf("正しくないデータです %#v\n", issue)
		return
	}

	// データベースに保存
	newJsonData, err := json.Marshal(issue)
	if err != nil {
		fmt.Printf("JSONデータを作成できませんでした。 %#v\n", err)
		return
	}
	if _, err = db.Exec("INSERT INTO issue (properties) VALUES ($1)", string(newJsonData)); err != nil {
		fmt.Println("データを保存できませんでした")
		return
	}

	fmt.Println("データを保存しました")
}
