curl -l -H "Content-type: application/json" -X POST -d \
'{
    "TaskID":"test_task_2", 
    "Transactions":[
        { "FromEnterprise": "kaka4", "FromAccount": "0001", "ToEnterprise": "kaka4", "ToAccount": "2002", "Amount": "50", "Date": "20170328", "Time": "121000", "HashID":"abcdabcd1" },
        { "FromEnterprise": "kaka4", "FromAccount": "0004", "ToEnterprise": "kaka4", "ToAccount": "2004", "Amount": "40", "Date": "20170328", "Time": "121000", "HashID":"abcdabcd4" },
        { "FromEnterprise": "kaka4", "FromAccount": "2002", "ToEnterprise": "kaka4", "ToAccount": "0002", "Amount": "30", "Date": "20170328", "Time": "121000", "HashID":"abcdabcd2" },
        { "FromEnterprise": "kaka4", "FromAccount": "0005", "ToEnterprise": "kaka4", "ToAccount": "2004", "Amount": "50", "Date": "20170328", "Time": "121000", "HashID":"abcdabcd5" },
        { "FromEnterprise": "kaka4", "FromAccount": "0001", "ToEnterprise": "kaka4", "ToAccount": "2001", "Amount": "50", "Date": "20170328", "Time": "120000", "HashID":"abcdabcd1" },
        { "FromEnterprise": "kaka4", "FromAccount": "0003", "ToEnterprise": "kaka4", "ToAccount": "2004", "Amount": "30", "Date": "20170328", "Time": "121000", "HashID":"abcdabcd3" }
    ]
}' \
http://127.0.0.1:33000/transactions