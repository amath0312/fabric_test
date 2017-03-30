curl -l -H "Content-type: application/json" -X PUT -d \
'{
    "EnterpriseID": "kaka4",
    "Accounts": [
        { "id": "2001", "balance": 2100 },
        { "id": "2002", "balance": 2200 },
        { "id": "2004", "balance": 2400 }
    ]
}'  \
http://127.0.0.1:33000/enterprises