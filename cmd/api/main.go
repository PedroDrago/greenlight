package main

const (
	version = "1.0.0"
)

func main() {
	app, db := newApplication()
	defer db.Close()
	app.logger.Info("Connection with DB established", nil)
	err := app.serve()
	app.logger.Fatal(err, nil)
}
