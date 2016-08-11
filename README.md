### Yad is a Yandex.Disk Go REST client

**Yandex.Disk homepage:** https://disk.yandex.ru

**Yandex.Disk REST API documentation:** https://tech.yandex.ru/disk/api

**Not implemented features for now:**

* Files publishing
* Resource meta information management
* Latest uploaded files

**Examples:**

```
token := "oauth-token"
c := yad.NewClient(token)


// List all files in the root of Disk.
list, err := c.ListAll("/asp2")
if err != nil {
	panic(err)
}
for _, res := range list.Items {
	fmt.Print(res.Name)
	if res.Type == ResourceTypeDir {
		fmt.Println("/")
	} else {
		fmt.Println()
	}
}

// Upload file.
file, err := os.Open("local-file-name")
if err != nil {
	panic(err)
}
defer file.Close()
err = c.Upload("remote-file-name", file)
if err != nil {
	panic(err)
}

// Download file.
file, err := os.Create("local-file-name")
if err != nil {
	panic(err)
}
defer file.Close()

_, err = c.Download("remote-file-name", file)
if err != nil {
	panic(err)
}

// Delete directory.
link, err := c.Delete("remote-file-name", false)
if err != nil {
	panic(err)
}
// Check delete operation status.
status, err := c.OpStatus(link)
if err != nil {
	panic(err)
}
fmt.Println("status:", status)
```

**Copying:** This programm is released under the GNU General Public License version 3 or later, which is distributed in the COPYING file. You should have received a copy of the GNU General Public

**Author:** Viacheslav Chimishuk <vchimishuk@yandex.ru>
