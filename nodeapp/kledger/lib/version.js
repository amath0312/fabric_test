var hfc = require('fabric-client');
var utils = require('fabric-client/lib/utils.js');

var log4js = require('log4js');
var logger = log4js.getLogger('VERSION');


var config = require('../config.json');
var helper = require('./helper.js');

logger.setLevel('DEBUG');

module.exports.getChainVersion = function(callback) {
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
                fcn: "version",
                args: []
            };
            // Query chaincode
            return chain.queryByChaincode(request);
        }
    ).then(
        function(response_payloads) {
            logger.debug(JSON.stringify(response_payloads));
            for (let i = 0; i < response_payloads.length; i++) {
                logger.info('############### Enterprise: \r\n    %s', response_payloads[i].toString('utf8'));
                if (i == 0) {
                    callback(response_payloads[i].toString('utf8'), null);
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