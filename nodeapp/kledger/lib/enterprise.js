var hfc = require('fabric-client');
var utils = require('fabric-client/lib/utils.js');

var log4js = require('log4js');
var logger = log4js.getLogger('ENTERPRISE');


var config = require('../config.json');
var helper = require('./helper.js');

logger.setLevel('DEBUG');

module.exports.createEnterprise = function(enterpriseId, accounts, callback) {
    var result = helper.getChain("chain_" + process.uptime, true);
    var client = result.client;
    var chain = result.chain;
    var eventhub = result.eventhub;
    var tx_id;
    hfc.newDefaultKeyValueStore({
        path: config.keyValueStore
    }).then(function(store) {
        client.setStateStore(store);
        return helper.getSubmitter(client);
    }).then(
        function(admin) {
            logger.info('Successfully obtained user to initial Enterprise.');
            logger.info('Executing transaction');
            tx_id = helper.getTxId();
            var nonce = utils.getNonce();
            var args = []
            args.push(enterpriseId);
            accounts.forEach(function(account) {
                args.push(account.id, account.balance.toString());
            });

            // send proposal to endorser
            var request = {
                chaincodeId: config.chaincodeID,
                fcn: "createEnterprise",
                args: args,
                chainId: config.channelID,
                txId: tx_id,
                nonce: nonce
            };
            return chain.sendTransactionProposal(request);
        }
    ).then(
        function(results) {
            logger.info('Successfully obtained proposal responses from endorsers');

            return helper.processProposal(chain, results, 'move');
        }
    ).then(
        function(response) {
            if (response.status === 'SUCCESS') {
                var handle = setTimeout(() => {
                    logger.error('Failed to receive transaction notification within the timeout period');
                    callback("1111", null);
                }, parseInt(config.waitTime));

                eventhub.registerTxEvent(tx_id.toString(), (tx) => {
                    logger.info('The chaincode transaction has been successfully committed, ' + tx_id.toString());
                    clearTimeout(handle);
                    eventhub.disconnect();
                    callback("0000", null);
                });
            } else {
                callback("9999", null);
            }
        }
    ).catch(
        function(err) {
            eventhub.disconnect();
            logger.error('Failed to invoke transaction due to error: ' + err.stack ? err.stack : err);
            callback("9999", err);
        }
    );
}
module.exports.addAccounts = function(enterpriseId, accounts, callback) {
    var result = helper.getChain("chain_" + process.uptime, true);
    var client = result.client;
    var chain = result.chain;
    var eventhub = result.eventhub;
    var tx_id;
    hfc.newDefaultKeyValueStore({
        path: config.keyValueStore
    }).then(function(store) {
        client.setStateStore(store);
        return helper.getSubmitter(client);
    }).then(
        function(admin) {
            logger.info('Successfully obtained user to initial Enterprise.');
            logger.info('Executing transaction');
            tx_id = helper.getTxId();
            var nonce = utils.getNonce();
            var args = []
            args.push(enterpriseId);
            accounts.forEach(function(account) {
                args.push(account.id, account.balance.toString());
            });

            // send proposal to endorser
            var request = {
                chaincodeId: config.chaincodeID,
                fcn: "addAccounts",
                args: args,
                chainId: config.channelID,
                txId: tx_id,
                nonce: nonce
            };
            return chain.sendTransactionProposal(request);
        }
    ).then(
        function(results) {
            logger.info('Successfully obtained proposal responses from endorsers');

            return helper.processProposal(chain, results, 'move');
        }
    ).then(
        function(response) {
            if (response.status === 'SUCCESS') {
                var handle = setTimeout(() => {
                    logger.error('Failed to receive transaction notification within the timeout period');
                    callback("1111", null);
                }, parseInt(config.waitTime));

                eventhub.registerTxEvent(tx_id.toString(), (tx) => {
                    logger.info('The chaincode transaction has been successfully committed, ' + tx_id.toString());
                    clearTimeout(handle);
                    eventhub.disconnect();
                    callback("0000", null);
                });
            } else {
                callback("9999", null);
            }
        }
    ).catch(
        function(err) {
            eventhub.disconnect();
            logger.error('Failed to invoke transaction due to error: ' + err.stack ? err.stack : err);
            callback("9999", err);
        }
    );
}
module.exports.getEnterprise = function(enterpriseId, callback) {
    var result = helper.getChain("chain_" + process.uptime, false);
    var client = result.client;
    var chain = result.chain;

    hfc.newDefaultKeyValueStore({
        path: config.keyValueStore
    }).then(function(store) {
        client.setStateStore(store);
        return helper.getSubmitter(client);
    }).then(
        function(admin) {
            logger.info('Successfully obtained enrolled user to perform query');

            logger.info('Executing Query');
            var targets = [];
            for (var i = 0; i < config.peers.length; i++) {
                targets.push(config.peers[i]);
            }
            //chaincode query request
            var request = {
                targets: targets,
                chaincodeId: config.chaincodeID,
                chainId: config.channelID,
                txId: utils.buildTransactionID(),
                nonce: utils.getNonce(),
                fcn: "getEnterprise",
                args: [enterpriseId]
            };

            return chain.queryByChaincode(request);
        }
    ).then(
        function(response_payloads) {
            logger.debug(JSON.stringify(response_payloads));
            for (let i = 0; i < response_payloads.length; i++) {
                logger.info('############### Enterprise: \r\n    %s', response_payloads[i].toString('utf8'));
                if (i == 0) {
                    var jsonData = response_payloads[i].toString('utf8');
                    if (jsonData == "") {
                        callback(null, null);
                    } else {
                        callback(JSON.parse(jsonData), null);
                    }
                }
            }
        }
    ).catch(
        function(err) {
            logger.error('error: ' + err.stack ? err.stack : err);
            callback(null, err);
        }
    );
}