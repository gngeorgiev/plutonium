var app = require('app');  // Module to control application life.
var BrowserWindow = require('browser-window');  // Module to create native browser window.

var config = require('./config');

require('crash-reporter').start();

var win = null;

app.on('window-all-closed', function () {
    if (process.platform != 'darwin') {
        app.quit();
    }
});

app.on('ready', function () {
    win = new BrowserWindow({
        width: 800,
        height: 600,
        icon: config.icon,
        title: config.name
    });

    win.loadUrl(config.url);

    var webContents = win.webContents;
    webContents.on('dom-ready', function () {
        webContents.executeJavaScript("require('electron-cookies')");
    });

    win.on('closed', function () {
        win = null;
    });
});
