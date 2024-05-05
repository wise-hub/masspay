```
fixed  lenght fields;
(1) (2) (3) (4)
-------------------------------------------------- -----------------------------------------
{ХХХХХХХХХХХ} {123456789.00} {First name} {Surname Surname} {Reason}
-------------------------------------------------- -----------------------------------------
(1) - 1 to 22 - Recipient account (IBAN)
(2) - 24 to 36 - Payment amount (right aligned, with decimal point, no leading zeros)
(3) - 38 to 72 - Recipient name
(4) - 74 tp 108 - Basis for payment

Notes:
- The data of each line should start from the first position
- One entry (line) is one payment
- The file name (for bulk payment) must start with the Latin letter O followed by the date (which will
mass payment is made) in the format DDMMYY, the Latin letter P and with the extension DPN, where N is
the next mass payment for the day (counting starts from 0). For example, first payment for date 26.05.2006 –
o260506p.dp0
- In the original file, there is a new line after the last record, also line width is 145 characters


// account := strings.TrimSpace(line[:22])
// amountStr := strings.TrimSpace(line[23:36])
// recipient := strings.TrimSpace(line[37:72])
// reason := strings.TrimSpace(line[73:108])


```


```
nginx

	upstream masspay_upstream {
		server localhost:9001;
		server localhost:9002;
		keepalive 128;
	}

    location /masspay {
        proxy_pass http://masspay_upstream/masspay;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection '';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }

```