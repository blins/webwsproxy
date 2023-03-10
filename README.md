# webwsproxy

HTTP to websocket proxy for message

## Использование

Вебсокет и управляющий http сервер биндятся на разные порты. Это сделано для того чтобы можно было разграничить доступ из reverse-proxy и из внутренних микросервисов.

Для получения сообщений надо подключиться к websocket по адресу ws://domail.tld:8080/websocket_entrypoint?channel=channelname

  * domail.tld - домен по которому доступна программа
  * websocket_entrypoint - указывается в настройках. По умолчанию равен '/ws'
  * channelname - обязательный аргумент, название канала по которому слушается сообщение (можно использовать несколько имен через запятую)

Для отправки надо в запросе GET или POST отправить два параметра на http://domail.tld:8000
  * 'channel' - название канала в который отправляется сообщение
  * 'msg' - сообщение которое передается (может быть любым. Передается без изменений, кроме url_decode)

## docker

Sample:
```bash
# run on default settings
docker run -d --name wsproxy blins1999/webwsproxy:latest 

#open in web browser http://wsproxy:8080
curl "http://wsproxy:8000/?channel=test&msg=hello%20from%20commandline"

# get help of usage
docker run -it --rm blins1999/webwsproxy:latest -h
```
