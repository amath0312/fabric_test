curl -l -H "Content-type: application/json" -X POST -d \
'{
    "EnterpriseID": "kaka4",
    "Accounts": [{
            "id": "0001",
            "balance": 100
        },
        {
            "id": "0002",
            "balance": 170
        },
        {
            "id": "0003",
            "balance": 300
        },
        {
            "id": "0004",
            "balance": 400
        },
        {
            "id": "0005",
            "balance": 500
        }
    ]
}'  \
http://127.0.0.1:33000/enterprises