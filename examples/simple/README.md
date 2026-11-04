# simple

Winit風のMado APIをSDL3バックエンドで使う最小例です。

実行:

```text
go run ./examples/simple
```

アプリケーションの `ApplicationHandler` はWinitバックエンド版と同じです。Window driverを `windowsdriver.NewFactory()` で作成し、`mado.NewWindow(factory, attributes)` で実ウィンドウを初期化します。

`main.go`にはSDL3とWinitの両方のインポートを記載しています。SDL3側を有効にしたまま、Winit側をコメントアウト解除する場合は、SDL3側をコメントアウトしてください。Factory名はどちらも`windowsdriver`なので、アプリケーション本体を変更せずにWindow driverだけを交換できます。

SDL3はSDLのイベント・入力デバイスやSDL GPUとの統合に向いています。WinitはWinitのネイティブなWindowライフサイクルを利用でき、WGPUとの組み合わせに向いています。

ウィンドウ作成時にバックエンドセッションも生成されます。マウス、キーボード、ジョイスティックは `window.Input()` から得られる入力Facadeを通して取得し、同じセッションを共有します。
