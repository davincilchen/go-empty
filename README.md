# DB

* Use MariaDB 10.7

## Create DB and Migrate

* 依序執行

* 1.createDB.exe
* 2.migrate.exe
* 3.loadSeed.exe

# App 註冊
* 圖片存放路徑: web/static 資料夾下
* DB註冊路徑: static/執行檔名稱 [欄位:image_url]

# 清理不要的資料夾後
* 變更專案名稱
* vscode設定
* 移除不要的cmd內資料夾
* 移除Business範例相關之資料夾
* 執行go mod tidy
* 砍掉readme.me此檔
# 專案分佈
* cmd下可建置其他執行檔
* pkg下app外層是上層通用的,含通用model (不含)deliverymodel
* pkg下app內層ctxcache,deliverymodel,errordef是內層通用
* pkg下app內層server內含進入點
* pkg下app內層含delivery, usecase , repo(db orm, http,grpc ...)
* Business範例app/user,app/device,app/edge
* 查看[https://github.com/davincilchen/goinit]