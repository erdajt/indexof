# Screenshots

[readme](../README.md) · [usage](usage.md) · [config](configuration.md)

some screenshots of indexof browsing debian pages. click an image for the full size.

## Browsing

[![debian links in indexof](images/browser.png)](images/browser.png)

`j` / `k` to move, `enter` to follow a directory, `h` to go back. the footer shows the download destination, cleanup setting, and theme.

## Editor

[![a downloaded html document open in neovim](images/editor.png)](images/editor.png)

press `enter` on a file to download it and open it in your editor. this is neovim. close it to return to indexof.

## Downloads

[![five completed downloads](images/downloads.png)](images/downloads.png)

`d` starts a download. `p` opens this view. these five jobs are done; running jobs show their progress here. you can pause, resume, or cancel them and keep browsing.

`0 B` on the right means the total size wasn't known. the number on the left is what was actually downloaded.

## Settings

[![the settings panel with the tokyo theme selected](images/settings.png)](images/settings.png)

press `s`, choose an option with `j` / `k`, and hit `enter` to change it. changes save automatically. this screenshot has a 70 mb threshold; the default is 50 mb.

## Visual selection

[![visual mode active in a listing](images/visual-mode.png)](images/visual-mode.png)

press `v`, move to select a range, then `y` to copy the urls. `esc` cancels. the checkmarks next to files mean they're already downloaded.

## Clipboard

[![status message after yanking three items](images/yank-confirmation.png)](images/yank-confirmation.png)

after copying, indexof reports the number of urls and goes back to normal mode. the clipboard needs to be available in your terminal session.

the screenshots use `tokyo`. the font, background, and transparency are from the terminal setup.
