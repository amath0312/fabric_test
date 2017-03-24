/**
 * Copyright 2016 IBM All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */
// This is Sample end-to-end standalone program that focuses on exercising all
// parts of the fabric APIs in a happy-path scenario
'use strict';

var log4js = require('log4js');
var logger = log4js.getLogger('INVOKE');

var hfc = require('fabric-client');
var utils = require('fabric-client/lib/utils.js');
var Peer = require('fabric-client/lib/Peer.js');
var Orderer = require('fabric-client/lib/Orderer.js');
var EventHub = require('fabric-client/lib/EventHub.js');

var config = require('./config.json');
var helper = require('./helper.js');

logger.setLevel('DEBUG');

var client = new hfc();
var chain;
var eventhub;
var tx_id = null;




// initMarble("red", "red", "300", "zhaol");
// transferMarble('marble2', 'jerry');

readMarble('marble4');
// getMarblesByRange('color~name', 'color~namez');
// queryMarbles("{\"selector\":{\"docType\":\"marble\",\"owner\":\"tom\"}}");
// queryMarblesByOwner('zhaol');

function initMarble(marbleName, color, size, owner) {
    _init(true);

    hfc.newDefaultKeyValueStore({
        path: config.keyValueStore
    }).then(function(store) {
        client.setStateStore(store);
        return helper.getSubmitter(client);
    }).then(
        function(admin) {
            logger.info('Successfully obtained user to initial marble.');

            logger.info('Executing transaction');
            tx_id = helper.getTxId();
            var nonce = utils.getNonce();
            // send proposal to endorser
            var request = {
                chaincodeId: config.chaincodeID,
                fcn: "initMarble",
                args: [marbleName, color, size, owner],
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
    ).then(_processProposalResult).catch(
        function(err) {
            eventhub.disconnect();
            logger.error('Failed to invoke transaction due to error: ' + err.stack ? err.stack : err);
        }
    );
}



function transferMarble(marbleName, newOwner) {
    _init(true);
    hfc.newDefaultKeyValueStore({
        path: config.keyValueStore
    }).then(function(store) {
        client.setStateStore(store);
        return helper.getSubmitter(client);
    }).then(
        function(admin) {
            logger.info('Successfully obtained user to transfer marble. executing transaction.');

            tx_id = helper.getTxId();
            var nonce = utils.getNonce();
            // send proposal to endorser
            var request = {
                chaincodeId: config.chaincodeID,
                fcn: "transferMarble",
                args: [marbleName, newOwner],
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
    ).then(_processProposalResult).catch(
        function(err) {
            eventhub.disconnect();
            logger.error('Failed to invoke transferMarble due to error: ' + err.stack ? err.stack : err);
        }
    );
}

function transferMarblesBasedOnColor(color, newOwner) {
    _init(true);
}

function deleteMarble(marbleName) {
    _init(true);
}



function readMarble(marbleName) {
    _init(false);
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
            var args = helper.getArgs(config.queryRequest.args);
            //chaincode query request
            var request = {
                targets: targets,
                chaincodeId: config.chaincodeID,
                chainId: config.channelID,
                txId: utils.buildTransactionID(),
                nonce: utils.getNonce(),
                fcn: "readMarble",
                args: [marbleName]
            };
            // Query chaincode
            return chain.queryByChaincode(request);
        }
    ).then(
        function(response_payloads) {
            for (let i = 0; i < response_payloads.length; i++) {
                logger.info('############### Marble: \r\n    %s', response_payloads[i].toString('utf8'));
            }
        }
    ).catch(
        function(err) {
            logger.error('Failed to end to end test with error:' + err.stack ? err.stack : err);
        }
    );
}

function getMarblesByRange(start, end) {
    _init(false);
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
                fcn: "getMarblesByRange",
                args: [start, end]
            };
            // Query chaincode
            return chain.queryByChaincode(request);
        }
    ).then(
        function(response_payloads) {
            for (let i = 0; i < response_payloads.length; i++) {
                logger.info('############### Marble: \r\n    %s', response_payloads[i].toString('utf8'));
            }
        }
    ).catch(
        function(err) {
            logger.error('Failed to end to end test with error:' + err.stack ? err.stack : err);
        }
    );
}

function queryMarblesByOwner(owner) {
    _init(false);
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
            var args = helper.getArgs(config.queryRequest.args);
            //chaincode query request
            var request = {
                targets: targets,
                chaincodeId: config.chaincodeID,
                chainId: config.channelID,
                txId: utils.buildTransactionID(),
                nonce: utils.getNonce(),
                fcn: "queryMarblesByOwner",
                args: [owner]
            };
            // Query chaincode
            return chain.queryByChaincode(request);
        }
    ).then(
        function(response_payloads) {
            for (let i = 0; i < response_payloads.length; i++) {
                logger.info('############### Marble: \r\n    %s', response_payloads[i].toString('utf8'));
            }
        }
    ).catch(
        function(err) {
            logger.error('Failed to end to end test with error:' + err.stack ? err.stack : err);
        }
    );
}

function queryMarbles(sql) {
    _init(false);
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
            var args = helper.getArgs(config.queryRequest.args);
            //chaincode query request
            var request = {
                targets: targets,
                chaincodeId: config.chaincodeID,
                chainId: config.channelID,
                txId: utils.buildTransactionID(),
                nonce: utils.getNonce(),
                fcn: "queryMarbles",
                args: [sql]
            };
            // Query chaincode
            return chain.queryByChaincode(request);
        }
    ).then(
        function(response_payloads) {
            for (let i = 0; i < response_payloads.length; i++) {
                logger.info('############### Marble: \r\n    %s', response_payloads[i].toString('utf8'));
            }
        }
    ).catch(
        function(err) {
            logger.error('Failed to end to end test with error:' + err.stack ? err.stack : err);
        }
    );
}




function _init(subscribeEvent) {
    chain = client.newChain(config.chainName);
    chain.addOrderer(new Orderer(config.orderer.orderer_url));
    if (subscribeEvent) {
        eventhub = new EventHub();
        eventhub.setPeerAddr(config.events[0].event_url);
        eventhub.connect();
    }
    for (var i = 0; i < config.peers.length; i++) {
        chain.addPeer(new Peer(config.peers[i].peer_url));
    }
}

function _processProposalResult(response) {
    if (response.status === 'SUCCESS') {
        var handle = setTimeout(() => {
            logger.error('Failed to receive transaction notification within the timeout period');
            process.exit(1);
        }, parseInt(config.waitTime));

        eventhub.registerTxEvent(tx_id.toString(), (tx) => {
            logger.info('The chaincode transaction has been successfully committed');
            clearTimeout(handle);
            eventhub.disconnect();
        });
    }
}