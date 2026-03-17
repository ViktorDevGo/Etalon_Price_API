# Email уведомления - Быстрый старт

## ✅ Что уже готово

Email сервис полностью реализован и интегрирован с синхронизацией номенклатуры.

## ⚠️ Что нужно сделать

### 1. Узнать IP сервера

```bash
curl -4 ifconfig.me
```

Запишите полученный IP адрес.

### 2. Добавить IP в whitelist SMTP

1. Войти в панель управления https://www.reg.ru/
2. Войти в свой аккаунт
3. Перейти в раздел "Почта"
4. Найти настройки безопасности / whitelist
5. Добавить IP адрес сервера для SMTP доступа
6. Сохранить

### 3. Проверить пароль

Проверьте что пароль `S69Y1ypojVLCZHO8` работает:
- Попробуйте войти на https://mail.hosting.reg.ru/
- Логин: admin@etalon-shina.ru
- Если не работает - сбросьте пароль и обновите в `.env`

### 4. Протестировать отправку

```bash
go run cmd/test-email/main.go
```

**Ожидаемый результат:**
```
✅ Test email sent successfully!
```

Проверьте почту v.boyarkin@etalon-shina.ru

### 5. Готово!

Теперь при каждой синхронизации будет автоматически отправляться email:

```bash
go run cmd/sync-nomenclature/main.go -type=tyres
```

## 📞 Если не работает

**Техподдержка reg.ru:**
- Email: support@reg.ru
- Телефон: 8 800 250-0-509

**Сообщите:**
- IP сервера: (результат curl -4 ifconfig.me)
- Логин: admin@etalon-shina.ru
- Ошибка: "connection reset by peer" при подключении к SMTP порт 587/465

## 📖 Подробная документация

Смотрите `docs/EMAIL_NOTIFICATIONS.md`
