# Termflow UI Testing Framework

Playwright風の端末アプリケーション自動テストフレームワークです。

## 概要

このフレームワークは、termflowベースのアプリケーションのUIを自動テストするためのツールです。実際の端末環境を模擬し、ユーザーインタラクションをシミュレートして、出力を検証できます。

## 機能

### 1. **TerminalTest** - 仮想端末テスト
- 実際のPTY（pseudo-terminal）を使用
- リアルなキーボード入力シミュレーション
- ANSI エスケープシーケンス対応
- スピナーアニメーション検証

### 2. **MockIO** - 軽量単体テスト
- I/O をモックしてコンポーネント単位でテスト
- 高速実行
- 出力フォーマット検証

## 使い方

### 基本的なテスト

```go
func TestMyApp(t *testing.T) {
    // アプリケーションを起動
    tt, err := NewTerminalTest(t, "./my-app", "--termflow")
    if err != nil {
        t.Fatalf("Failed to start: %v", err)
    }
    defer tt.Close()

    // 起動を待つ
    tt.Wait(500 * time.Millisecond)

    // ウェルカムメッセージを確認
    tt.ExpectWelcome()
    tt.ExpectPrompt()

    // ユーザー入力をシミュレート
    tt.Type("hello world")

    // 応答を待って検証
    tt.Wait(1 * time.Second)
    tt.ExpectOutput("Hello!")
}
```

### 高度なテスト

```go
func TestSpinnerAndCtrlC(t *testing.T) {
    tt, err := NewTerminalTest(t, "./app", "--termflow")
    if err != nil {
        t.Fatalf("Failed to start: %v", err)
    }
    defer tt.Close()

    tt.Wait(500 * time.Millisecond)

    // 長い処理をトリガー
    tt.Type("complex task")

    // スピナーアニメーションを確認
    tt.ExpectSpinner()
    tt.ExpectThinking()

    // Ctrl+C のテスト
    tt.SendCtrlC()
    tt.ExpectOutput("Press Ctrl+C again to exit")
    tt.ExpectNoCtrlC() // ^C文字が表示されないことを確認

    // 2回目のCtrl+C
    tt.SendCtrlC()
    tt.ExpectOutput("Goodbye!")
}
```

### マルチライン入力テスト

```go
func TestMultilineInput(t *testing.T) {
    tt, err := NewTerminalTest(t, "./app", "--termflow")
    if err != nil {
        t.Fatalf("Failed to start: %v", err)
    }
    defer tt.Close()

    tt.Wait(500 * time.Millisecond)

    // マルチライン入力をトリガー
    tt.SendKeys("Write a function...")
    tt.SendEnter()

    // 継続プロンプトを確認
    tt.ExpectOutput("Continue typing")
    tt.ExpectPattern(`\d+>`) // "2>" のような行番号プロンプト

    // 追加の行を入力
    tt.SendKeys("def hello():")
    tt.SendEnter()
    tt.SendKeys("    print('Hello')")
    tt.SendEnter()

    // 終了マーカー
    tt.SendKeys(".")
    tt.SendEnter()

    tt.ExpectThinking()
}
```

## 利用可能なメソッド

### 入力操作
- `SendKeys(string)` - キー入力送信
- `Type(string)` - テキスト入力 + Enter
- `SendEnter()` - Enter キー
- `SendCtrlC()` - Ctrl+C
- `Wait(duration)` - 待機

### 出力検証
- `ExpectOutput(string)` - テキストが含まれることを確認
- `ExpectPattern(regex)` - 正規表現マッチを確認
- `ExpectPrompt()` - プロンプト（✦）の存在確認
- `ExpectWelcome()` - ウェルカムメッセージ確認
- `ExpectSpinner()` - スピナーアニメーション確認
- `ExpectThinking()` - "Thinking..." メッセージ確認
- `ExpectNoCtrlC()` - ^C 文字が表示されないことを確認

### デバッグ
- `GetOutput()` - 生出力取得（ANSIコード含む）
- `GetVisibleOutput()` - 表示テキスト取得（ANSIコード除去）
- `Screenshot()` - 現在の画面状態をフォーマット表示
- `GetLines()` - 行ごとの出力配列

## テスト実行

```bash
# 基本テスト
go test ./lib/termflow/testing

# 統合テスト（バイナリビルド必要）
go build -o /tmp/rigel-test cmd/rigel/main.go
go test ./lib/termflow/testing -v

# ベンチマーク
go test -bench=. ./lib/termflow/testing

# 短時間テストのみ（統合テストをスキップ）
go test -short ./lib/termflow/testing
```

## アーキテクチャ

```
┌─────────────────────────────────────┐
│           Test Code                 │
├─────────────────────────────────────┤
│      TerminalTest Framework         │
├─────────────────────────────────────┤
│    PTY (Pseudo Terminal)            │
├─────────────────────────────────────┤
│    Target Application               │
│    (Rigel with --termflow)          │
└─────────────────────────────────────┘
```

## 制限事項

- Linux/macOS での PTY サポートが必要
- Windows では制限あり（WSL推奨）
- アプリケーションのビルドが必要
- タイミングに依存するテストは不安定になる可能性

## ベストプラクティス

1. **適切な待機時間**: `Wait()` で十分な時間を設ける
2. **スクリーンショット活用**: デバッグ時は `Screenshot()` を使用
3. **段階的検証**: 大きな動作を小さなステップに分けて検証
4. **環境の初期化**: 各テストで新しいTerminalTestインスタンス使用
5. **エラーハンドリング**: defer で必ずCloseを呼ぶ

このフレームワークにより、termflowアプリケーションの品質を自動テストで保証できます。
