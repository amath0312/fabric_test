var express = require('express');
var app = express();
var ent = require('./lib/enterprise.js');
var ver = require('./lib/version.js');
var trans = require('./lib/transaction.js');

var log4js = require('log4js');
var logger = log4js.getLogger('MAIN');
logger.setLevel('DEBUG');

app.get('/', function(req, res) {
    res.json({ "result": "0000", "message": "hello" });
})

app.get('/enterprises/:enterpriseid', function(req, res) {
    var enterpriseid = req.params.enterpriseid;
    ent.getEnterprise(enterpriseid, function(result, err) {
        if (err == null) {
            enterprise = result;
            if (enterprise == null) {
                res.json({ "result": "0000", "message": "success", "data": null });
            } else {
                var jdata = { "result": "0000", "message": "success", "data": enterprise };
                res.json(jdata);
            }
        } else {
            logger.error('===================:' + err.stack ? err.stack : err);
            res.json({ "result": "9999", "message": "create enterprise error: " + (err.stack ? err.stack : err) });
        }
    })
})

app.post('/enterprises', function(req, res) {
    req.on("data", function(data) {
        var enterprise = JSON.parse(data.toString());
        ent.createEnterprise(enterprise.EnterpriseID, enterprise.Accounts, function(result, err) {
            if (err == null) {
                res.json({ "result": "0000", "message": "success" });
            } else {
                res.json({ "result": "9999", "message": "create enterprise error: " + (err.stack ? err.stack : err) });
            }
        });
    });
})

app.put('/enterprises', function(req, res) {
    req.on("data", function(data) {
        var enterprise = JSON.parse(data.toString());
        ent.addAccounts(enterprise.EnterpriseID, enterprise.Accounts, function(result, err) {
            if (err == null) {
                res.json({ "result": "0000", "message": "success" });
            } else {
                res.json({ "result": "9999", "message": "add accounts to enterprise error: " + (err.stack ? err.stack : err) });
            }
        });
    });
})

app.post('/transactions', function(req, res) {
    req.on("data", function(data) {
        var transaction = JSON.parse(data.toString());
        console.log(transaction);
        trans.makeTransactions(transaction.TaskID, transaction.Transactions, function(result, err) {
            if (err == null) {
                res.json({ "result": "0000", "message": "success" });
            } else {
                res.json({ "result": "9999", "message": "create enterprise error: " + (err.stack ? err.stack : err) });
            }
        });
    });
})

app.get('/tasks/:taskid', function(req, res) {
    var taskid = req.params.taskid;
    trans.getTask(taskid, function(result, err) {
        if (err == null) {
            task = result;
            if (task == null) {
                res.json({ "result": "0000", "message": "success", "data": null });
            } else {
                var jdata = { "result": "0000", "message": "success", "data": task };
                res.json(jdata);
            }
        } else {
            logger.error('===================:' + err.stack ? err.stack : err);
            res.json({ "result": "9999", "message": "create enterprise error: " + (err.stack ? err.stack : err) });
        }
    })
})

// http://127.0.0.1:33000/transactions/20170329_4229491576537700
app.get('/transactions/:transactionid', function(req, res) {
    var transactionid = req.params.transactionid;
    trans.getTransaction(transactionid, function(result, err) {
        if (err == null) {
            transaction = result;
            if (transaction == null) {
                res.json({ "result": "0000", "message": "success", "data": null });
            } else {
                var jdata = { "result": "0000", "message": "success", "data": transaction };
                res.json(jdata);
            }
        } else {
            logger.error('===================:' + err.stack ? err.stack : err);
            res.json({ "result": "9999", "message": "create enterprise error: " + (err.stack ? err.stack : err) });
        }
    })
})

// http://127.0.0.1:33000/transactions?date[from]=20170329&date[to]=20170329
// http://127.0.0.1:33000/transactions?enterpriseId=kaka4&accountId=2001&date=20170328
app.get('/transactions', function(req, res) {
    var enterpriseId = req.query.enterpriseId;
    console.log(enterpriseId);
    if (enterpriseId != null && enterpriseId != "") {
        // query by enterpriseid, account, date
        var accountId = req.query.accountId;
        var date = req.query.date;
        trans.queryTransactions(enterpriseId, accountId, date, function(result, err) {
            if (err == null) {
                transaction = result;
                if (transaction == null) {
                    res.json({ "result": "0000", "message": "success", "data": null });
                } else {
                    var jdata = { "result": "0000", "message": "success", "data": transaction };
                    res.json(jdata);
                }
            } else {
                logger.error('===================:' + err.stack ? err.stack : err);
                res.json({ "result": "9999", "message": "create enterprise error: " + (err.stack ? err.stack : err) });
            }
        })
    } else {
        // query by date
        var from = req.query.date.from;
        var to = req.query.date.to;
        if (to == null || to == "") {
            to = "20991231";
        }

        trans.queryTransactionsByDate(from, to, function(result, err) {
            if (err == null) {
                transaction = result;
                if (transaction == null) {
                    res.json({ "result": "0000", "message": "success", "data": null });
                } else {
                    var jdata = { "result": "0000", "message": "success", "data": transaction };
                    res.json(jdata);
                }
            } else {
                logger.error('===================:' + err.stack ? err.stack : err);
                res.json({ "result": "9999", "message": "create enterprise error: " + (err.stack ? err.stack : err) });
            }
        })
    }
})



app.get('/version', function(req, res) {
    ver.getChainVersion(function(result, err) {
        if (err == null) {
            logger.debug(result);
            res.send(result);
        } else {
            logger.error('===================:' + err.stack ? err.stack : err);
            res.send(err.stack ? err.stack : err);
        }
    })
})

app.listen(33000);