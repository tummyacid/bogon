package main

import (
	"log"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/nickvanw/bogon"
	"github.com/nickvanw/bogon/commands/config"
	"github.com/nickvanw/bogon/commands/config/viperprovider"
	"github.com/nickvanw/bogon/commands/plugins"
	"github.com/nickvanw/bogon/commands/plugins/bookmarks"
	"github.com/nickvanw/bogon/commands/plugins/bookmarks/boltdb"
	"github.com/nickvanw/bogon/commands/plugins/twitter"
	"github.com/nickvanw/ircx"
	"github.com/spf13/viper"
)

func main() {
	app := cli.NewApp()
	app.Name = "bogon"
	app.Action = realMain
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "name, n",
			Value:  "ircx",
			Usage:  "IRC Name to use",
			EnvVar: "BOGON_NAME",
		},
		cli.StringFlag{
			Name:   "server, s",
			Value:  "swarm-node02.nvw.io:6667",
			Usage:  "IRC host:port to connect to",
			EnvVar: "BOGON_SERVER",
		},
		cli.StringFlag{
			Name:   "user, u",
			Value:  "",
			Usage:  "User in User:Pass to use when connecting, must have both",
			EnvVar: "BOGON_USER",
		},
		cli.StringFlag{
			Name:   "pass, p",
			Value:  "",
			Usage:  "Password in User:Pass to use when connecting, must have both",
			EnvVar: "BOGON_PASSWORD",
		},
		cli.StringFlag{
			Name:   "channels, c",
			Usage:  "channels to join on startup",
			Value:  "#test",
			EnvVar: "BOGON_CHANNELS",
		},
		cli.StringFlag{
			Name:   "config, cfg",
			Usage:  "config file",
			Value:  "",
			EnvVar: "BOGON_CONFIG",
		},
		cli.StringFlag{
			Name:   "bookmark, bm",
			Usage:  "location of bookmark db.",
			Value:  "bm.bdb",
			EnvVar: "BOGON_BOOKMARK_FILE",
		},
		cli.StringFlag{
			Name:   "admin, a",
			Usage:  "location of admin socket",
			Value:  "",
			EnvVar: "BOGON_ADMIN_SOCKET",
		},
	}
	app.Run(os.Args)
}

func realMain(c *cli.Context) {
	// Create the underlying connection
	bot := ircx.Classic(c.String("server"), c.String("name"))
	bot.Config.MaxRetries = 10
	if usr := c.String("user"); usr != "" {
		bot.Config.User = usr
	}
	bot.Config.Password = c.String("pass")

	channels := strings.Split(c.String("channels"), ",")

	// Create a new bogon
	bogon, err := bogon.New(bot, c.String("name"), channels)
	if err != nil {
		log.Fatalf("Unable to start new bogon: %s", err)
	}

	// Setup & add config
	viper.AutomaticEnv()

	// Add config file if provided
	if cfg := c.String("config"); cfg != "" {
		viper.SetConfigFile(cfg)
		viper.WatchConfig()
		if err := viper.ReadInConfig(); err != nil {
			log.Fatalf("Unable to read config: %s", err)
		}
	}

	// Add viper provider
	config.RegisterProvider(viperprovider.V{})

	// Set up commands
	if err := commandSetup(bogon, c); err != nil {
		log.Fatalf("error setting up commands: %s", err)
	}

	if cfg := c.String("admin"); cfg != "" {
		bogon.AdminSocket(cfg)
	}

	// Connect!
	if err := bogon.Connect(); err != nil {
		log.Fatalf("Unable to connect: %s", err)
	}
	bogon.Start()
}

func commandSetup(bogon *bogon.Client, c *cli.Context) error {
	// Register basic plugins
	bogon.AddCommands(plugins.Exports()...)

	// Register twitter command
	api, err := twitter.NewFromEnv()
	if err != nil && err != twitter.ErrMissingTokens {
		return err
	}
	bogon.AddCommands(api.TwitterHandler())

	// Register bookmark handler
	if bmdb := c.String("bookmark"); bmdb != "" {
		db, err := boltdb.New(bmdb)
		if err != nil {
			return err
		}
		bm := bookmarks.New(db)
		bogon.AddCommands(bm.Exports()...)
		bm.Block(bogon.ListCommands())
	}

	return nil
}
