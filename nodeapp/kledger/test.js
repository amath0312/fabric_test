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
var helper = require('../helper.js');

logger.setLevel('DEBUG');

var client = new hfc();
var chain;
var eventhub;
var tx_id = null;

// getChaincodeVersion();

// createEnterprice(
//     'kaka', [
//         { "id": "0001", "balance": 100 },
//         { "id": "0002", "balance": 200 },
//         { "id": "0003", "balance": 300 },
//         { "id": "0004", "balance": 400 },
//         { "id": "0005", "balance": 500 }
//     ]
// );

getEnterprice("kaka");

function getChaincodeVersion() {
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
                fcn: "version",
                args: []
            };
            // Query chaincode
            return chain.queryByChaincode(request);
        }
    ).then(
        function(response_payloads) {
            for (let i = 0; i < response_payloads.length; i++) {
                logger.info('############### version: %s', response_payloads[i].toString('utf8'));
            }
        }
    ).catch(
        function(err) {
            logger.error('Failed to end to end test with error:' + err.stack ? err.stack : err);
        }
    );
}


function createEnterprice(enterpriceId, accounts) {
    _init(true);

    hfc.newDefaultKeyValueStore({
        path: config.keyValueStore
    }).then(function(store) {
        client.setStateStore(store);
        return helper.getSubmitter(client);
    }).then(
        function(admin) {
            logger.info('Successfully obtained user to initial enterprice.');
            logger.info('Executing transaction');
            tx_id = helper.getTxId();
            var nonce = utils.getNonce();
            var args = []
            args.push(enterpriceId);
            accounts.forEach(function(account) {
                args.push(account.id, account.balance.toString());
            });

            // send proposal to endorser
            var request = {
                chaincodeId: config.chaincodeID,
                fcn: "createEnterprice",
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
    ).then(_processProposalResult).catch(
        function(err) {
            eventhub.disconnect();
            logger.error('Failed to invoke transaction due to error: ' + err.stack ? err.stack : err);
        }
    );
}


function getEnterprice(enterpriceId) {
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
                fcn: "getEnterprice",
                args: [enterpriceId]
            };
            // Query chaincode
            return chain.queryByChaincode(request);
        }
    ).then(
        function(response_payloads) {
            for (let i = 0; i < response_payloads.length; i++) {
                logger.info('############### enterprice: \r\n    %s', response_payloads[i].toString('utf8'));
            }
        }
    ).catch(
        function(err) {
            logger.error('Failed to end to end test with error:' + err.stack ? err.stack : err);
        }
    );
}


function _init(subscribeEvent) {
    if (chain) {
        return;
    }
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