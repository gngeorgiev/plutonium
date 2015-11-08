package main

import (
	"bytes"
	"github.com/codegangsta/cli"
	"github.com/kardianos/osext"
	"github.com/termie/go-shutil"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type AppConfig struct {
	Name, Url, Icon string
}

var pwd string

func main() {
	dir, err := osext.ExecutableFolder()
	if err != nil {
		log.Println(err)
		return
	}

	pwd = dir

	app := cli.NewApp()
	app.Name = "plutonium"
	app.Usage = "create true desktop applications from web pages"
	app.Version = "0.1"

	app.Commands = []cli.Command{
		getCreateCommand(),
	}

	app.Run(os.Args)
}

func exists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

func download(resourceUrl, dest string) error {
	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer file.Close()

	c := http.Client{}
	resp, err := c.Get(resourceUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func createApp(c *cli.Context) {
	name := c.String("name")
	appUrl := c.String("url")
	icon := c.String("icon")

	templatePath := filepath.Join(pwd, "data", "template")
	appPath := filepath.Join(pwd, "Applications", name)

	var err error
	//copy the template to the new app dir
	if exists(appPath) {
		log.Println("App exists, deleting it.")
		err = os.RemoveAll(appPath)
		if err != nil {
			log.Println(err)
			return
		}
	}

	log.Println("Copying template...")
	err = shutil.CopyTree(templatePath, appPath, nil)
	if err != nil {
		log.Println(err)
		return
	}

	//copy the icon to the new app
	var appIcon string
	if strings.Index(icon, "http://") != -1 ||
		strings.Index(icon, "https://") != -1 {
		log.Println("Downloading icon...")
		ext := filepath.Ext(icon)
		appIcon = filepath.Join(appPath, name+ext)
		err = download(icon, appIcon)
		if err != nil {
			log.Println(err)
			return
		}
	} else {
		iconName := filepath.Base(icon)
		appIcon = filepath.Join(appPath, iconName)
		log.Println("Copying icon...")
		err = shutil.CopyFile(icon, appIcon, true)
		if err != nil {
			log.Println(err)
			return
		}
	}

	//create the config to pass the template config
	config := AppConfig{name, appUrl, appIcon}

	//parse the config template
	log.Println("Processing config...")
	appConfigTemplatePath := filepath.Join(appPath, "config.template.js")
	configTemplate, err := ioutil.ReadFile(appConfigTemplatePath)
	if err != nil {
		log.Println(err)
		return
	}

	t := template.New(config.Name)
	t, err = t.Parse(string(configTemplate))
	if err != nil {
		log.Println(err)
		return
	}

	var buffer bytes.Buffer
	err = t.Execute(&buffer, config)
	if err != nil {
		log.Println(err)
		return
	}

	//write the new config
	err = ioutil.WriteFile(filepath.Join(appPath, "config.js"), buffer.Bytes(), os.ModePerm)
	if err != nil {
		log.Println(err)
		return
	}

	//remove the config template file
	log.Println("Cleaning up...")
	err = os.Remove(appConfigTemplatePath)
	if err != nil {
		log.Println(err)
		return
	}

	//create the .desktop entry
	home := os.Getenv("HOME")
	desktopEntryPath := filepath.Join(home, ".local/share/applications", "plutonium-"+config.Name+".desktop")
	log.Println("Creating desktop entry in ", desktopEntryPath)
	var entryBuffer bytes.Buffer
	entryBuffer.WriteString("[Desktop Entry]\r\n")
	entryBuffer.WriteString("Version=" + c.App.Version + "\r\n")
	entryBuffer.WriteString("Name=" + config.Name + "\r\n")
	entryBuffer.WriteString("Icon=" + config.Icon + "\r\n")
	entryBuffer.WriteString("Terminal=false\r\n")
	entryBuffer.WriteString("Categories=Utility\r\n")
	entryBuffer.WriteString("Type=Application\r\n")
	entryBuffer.WriteString("NoDisplay=false\r\n")

	electronPath := filepath.Join(pwd, "data/electron/electron")
	execCmd := electronPath + " " + appPath
	entryBuffer.WriteString("Exec=" + execCmd + "\r\n")

	if exists(desktopEntryPath) {
		log.Println("Desktop entry exists, deleting...")
		err = os.RemoveAll(desktopEntryPath)
		if err != nil {
			log.Println(err)
			return
		}
	}

	err = ioutil.WriteFile(desktopEntryPath, entryBuffer.Bytes(), os.ModePerm)
	if err != nil {
		log.Println(err)
		return
	}

	log.Println("Done creating", config.Name, "!")
}

func getCreateCommand() cli.Command {
	return cli.Command{
		Name:    "create",
		Aliases: []string{"c"},
		Usage:   "Create a desktop application",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "name, n",
				Usage: "Set your app's name",
			},
			cli.StringFlag{
				Name:  "url, u",
				Usage: "Set your app's url",
			},
			cli.StringFlag{
				Name:  "icon, i",
				Usage: "Set your app's icon",
			},
		},
		Action: createApp,
	}
}
