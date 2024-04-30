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

type BugIssue struct {
	ID         int    `json:"id" validate:"required"`
	ReporterID int    `json:"reporter_id" validate:"required"`
	Priority   int    `json:"priority" validate:"required"`
	Status     string `json:"status" validate:"required"`
	BugOptions
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

	jsonDataList := []string{
		// Bug Issue 1
		`{
			"id": 1,
			"reporter_id": 11,
			"priority": 1,
			"status": "new",
			"severity": "low",
			"version_affected": "1.0.0"
		}`,
		// Bug Issue 2
		`{
			"id": 2,
			"reporter_id": 22,
			"priority": 2,
			"status": "new",
			"severity": "middle",
			"version_affected": "1.0.0"
		}`,
		// Feature Request Issue
		`{
			"id": 3,
			"reporter_id": 33,
			"priority": 3,
			"status": "new",
			"sponsor_id": 333
		}`,
		// Bug Issue 余分なデータを含む
		`{
			"id": 4,
			"reporter_id": 44,
			"priority": 4,
			"status": "new",
			"severity": "high",
			"version_affected": "1.0.0",
			"extra": "extra data"
		}`,
		// Bug Issue 必須カラム不足
		`{
			"id": 5,
			"reporter_id": 55,
			"priority": 5,
			"status": "new",
			"severity": "high"
		}`,
		// Bug Issue かつ Feature Request Issue
		`{
			"id": 6,
			"reporter_id": 66,
			"priority": 6,
			"status": "new",
			"severity": "high",
			"version_affected": "1.0.0",
			"sponsor_id": 666
		}`,
	}

	for _, jsonData := range jsonDataList {
		if _, err = db.Exec("INSERT INTO issue (properties) VALUES ($1)", jsonData); err != nil {
			fmt.Println("データを保存できませんでした")
			return
		}
	}

	// Issueテーブルからデータを取得
	rows, err := db.Query("SELECT properties FROM issue")
	if err != nil {
		fmt.Println("データを取得できませんでした:", err)
		return
	}
	defer rows.Close()

	bugIssueList := []BugIssue{}
	for rows.Next() {
		var jsonData string
		if err := rows.Scan(&jsonData); err != nil {
			fmt.Println("データを取得できませんでした:", err)
			continue
		}

		// 取得したデータをIssue構造体に変換
		var issue Issue
		if err := json.Unmarshal([]byte(jsonData), &issue); err != nil {
			fmt.Println("データを変換できませんでした:", jsonData)
			continue
		}

		// バリデーション
		validate := validator.New()
		if err := validate.Struct(issue); err != nil {
			fmt.Printf("バリデーションエラーです。 %#v\n", issue)
			continue
		}

		if isBugIssue(issue) == isFeatureRequestIssue(issue) {
			fmt.Printf("正しくないデータです %#v\n", issue)
			continue
		}

		if !isBugIssue(issue) {
			continue
		}

		bugIssueList = append(bugIssueList, BugIssue{
			ID:         issue.ID,
			ReporterID: issue.ReporterID,
			Priority:   issue.Priority,
			Status:     issue.Status,
			BugOptions: *issue.BugOptions,
		})
	}

	// Bug Issueのリストを表示
	for _, bugIssue := range bugIssueList {
		// Bug Issueに対して処理を行う
		// fmt.Printf("Bug Issue: %#v\n", bugIssue)
		fmt.Println("Bug Issue ID: ", bugIssue.ID)
	}
}
