package main

import (
	"github.com/k0kubun/pp"
	"github.com/kociumba/konfig"
)

type MyData struct {
	Username string `toml:"username"`
	Password string `toml:"password"`
}

var (
	globalData MyData = MyData{
		Username: "username",
		Password: "password",
	}
)

func main() {
	mngr, err := konfig.NewKonfigManager(konfig.KonfigOptions{
		Format:       konfig.TOML,
		AutoLoad:     false,
		AutoSave:     true,
		UseCallbacks: true,
		KonfigPath:   "config.toml",
	})
	if err != nil {
		panic(err)
	}

	section := konfig.NewKonfigSection(&globalData,
		konfig.WithSectionName(func() string { return "global" }),
		konfig.WithOnLoad(func() error {
			pp.Println("OnLoad callback")
			return nil
		}),
	)

	mngr.RegisterSection(section)

	// this is overwritten by loading data from the file
	globalData.Username = "new_username"

	// due to go's limitations if your app closes naturally without an os signal you need to defer a call to Save() in you main()
	defer func() {
		if err := mngr.Save(); err != nil {
			panic(err)
		}
	}()

	// loading the data before saving changes overwrites it
	if err := mngr.Load(); err != nil {
		panic(err)
	}

	pp.Print(globalData)
}
