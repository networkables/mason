// Copyright 2024 David Hallum. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package netflows

const (
	IPFIX_FIELD_Reserved                              = 0
	IPFIX_FIELD_octetDeltaCount                       = 1
	IPFIX_FIELD_packetDeltaCount                      = 2
	IPFIX_FIELD_deltaFlowCount                        = 3
	IPFIX_FIELD_protocolIdentifier                    = 4
	IPFIX_FIELD_ipClassOfService                      = 5
	IPFIX_FIELD_tcpControlBits                        = 6
	IPFIX_FIELD_sourceTransportPort                   = 7
	IPFIX_FIELD_sourceIPv4Address                     = 8
	IPFIX_FIELD_sourceIPv4PrefixLength                = 9
	IPFIX_FIELD_ingressInterface                      = 10
	IPFIX_FIELD_destinationTransportPort              = 11
	IPFIX_FIELD_destinationIPv4Address                = 12
	IPFIX_FIELD_destinationIPv4PrefixLength           = 13
	IPFIX_FIELD_egressInterface                       = 14
	IPFIX_FIELD_ipNextHopIPv4Address                  = 15
	IPFIX_FIELD_bgpSourceAsNumber                     = 16
	IPFIX_FIELD_bgpDestinationAsNumber                = 17
	IPFIX_FIELD_bgpNextHopIPv4Address                 = 18
	IPFIX_FIELD_postMCastPacketDeltaCount             = 19
	IPFIX_FIELD_postMCastOctetDeltaCount              = 20
	IPFIX_FIELD_flowEndSysUpTime                      = 21
	IPFIX_FIELD_flowStartSysUpTime                    = 22
	IPFIX_FIELD_postOctetDeltaCount                   = 23
	IPFIX_FIELD_postPacketDeltaCount                  = 24
	IPFIX_FIELD_minimumIpTotalLength                  = 25
	IPFIX_FIELD_maximumIpTotalLength                  = 26
	IPFIX_FIELD_sourceIPv6Address                     = 27
	IPFIX_FIELD_destinationIPv6Address                = 28
	IPFIX_FIELD_sourceIPv6PrefixLength                = 29
	IPFIX_FIELD_destinationIPv6PrefixLength           = 30
	IPFIX_FIELD_flowLabelIPv6                         = 31
	IPFIX_FIELD_icmpTypeCodeIPv4                      = 32
	IPFIX_FIELD_igmpType                              = 33
	IPFIX_FIELD_samplingInterval                      = 34
	IPFIX_FIELD_samplingAlgorithm                     = 35
	IPFIX_FIELD_flowActiveTimeout                     = 36
	IPFIX_FIELD_flowIdleTimeout                       = 37
	IPFIX_FIELD_engineType                            = 38
	IPFIX_FIELD_engineId                              = 39
	IPFIX_FIELD_exportedOctetTotalCount               = 40
	IPFIX_FIELD_exportedMessageTotalCount             = 41
	IPFIX_FIELD_exportedFlowRecordTotalCount          = 42
	IPFIX_FIELD_ipv4RouterSc                          = 43
	IPFIX_FIELD_sourceIPv4Prefix                      = 44
	IPFIX_FIELD_destinationIPv4Prefix                 = 45
	IPFIX_FIELD_mplsTopLabelType                      = 46
	IPFIX_FIELD_mplsTopLabelIPv4Address               = 47
	IPFIX_FIELD_samplerId                             = 48
	IPFIX_FIELD_samplerMode                           = 49
	IPFIX_FIELD_samplerRandomInterval                 = 50
	IPFIX_FIELD_classId                               = 51
	IPFIX_FIELD_minimumTTL                            = 52
	IPFIX_FIELD_maximumTTL                            = 53
	IPFIX_FIELD_fragmentIdentification                = 54
	IPFIX_FIELD_postIpClassOfService                  = 55
	IPFIX_FIELD_sourceMacAddress                      = 56
	IPFIX_FIELD_postDestinationMacAddress             = 57
	IPFIX_FIELD_vlanId                                = 58
	IPFIX_FIELD_postVlanId                            = 59
	IPFIX_FIELD_ipVersion                             = 60
	IPFIX_FIELD_flowDirection                         = 61
	IPFIX_FIELD_ipNextHopIPv6Address                  = 62
	IPFIX_FIELD_bgpNextHopIPv6Address                 = 63
	IPFIX_FIELD_ipv6ExtensionHeaders                  = 64
	IPFIX_FIELD_mplsTopLabelStackSection              = 70
	IPFIX_FIELD_mplsLabelStackSection2                = 71
	IPFIX_FIELD_mplsLabelStackSection3                = 72
	IPFIX_FIELD_mplsLabelStackSection4                = 73
	IPFIX_FIELD_mplsLabelStackSection5                = 74
	IPFIX_FIELD_mplsLabelStackSection6                = 75
	IPFIX_FIELD_mplsLabelStackSection7                = 76
	IPFIX_FIELD_mplsLabelStackSection8                = 77
	IPFIX_FIELD_mplsLabelStackSection9                = 78
	IPFIX_FIELD_mplsLabelStackSection10               = 79
	IPFIX_FIELD_destinationMacAddress                 = 80
	IPFIX_FIELD_postSourceMacAddress                  = 81
	IPFIX_FIELD_interfaceName                         = 82
	IPFIX_FIELD_interfaceDescription                  = 83
	IPFIX_FIELD_samplerName                           = 84
	IPFIX_FIELD_octetTotalCount                       = 85
	IPFIX_FIELD_packetTotalCount                      = 86
	IPFIX_FIELD_flagsAndSamplerId                     = 87
	IPFIX_FIELD_fragmentOffset                        = 88
	IPFIX_FIELD_forwardingStatus                      = 89
	IPFIX_FIELD_mplsVpnRouteDistinguisher             = 90
	IPFIX_FIELD_mplsTopLabelPrefixLength              = 91
	IPFIX_FIELD_srcTrafficIndex                       = 92
	IPFIX_FIELD_dstTrafficIndex                       = 93
	IPFIX_FIELD_applicationDescription                = 94
	IPFIX_FIELD_applicationId                         = 95
	IPFIX_FIELD_applicationName                       = 96
	IPFIX_FIELD_postIpDiffServCodePoint               = 98
	IPFIX_FIELD_multicastReplicationFactor            = 99
	IPFIX_FIELD_className                             = 100
	IPFIX_FIELD_classificationEngineId                = 101
	IPFIX_FIELD_layer2packetSectionOffset             = 102
	IPFIX_FIELD_layer2packetSectionSize               = 103
	IPFIX_FIELD_layer2packetSectionData               = 104
	IPFIX_FIELD_bgpNextAdjacentAsNumber               = 128
	IPFIX_FIELD_bgpPrevAdjacentAsNumber               = 129
	IPFIX_FIELD_exporterIPv4Address                   = 130
	IPFIX_FIELD_exporterIPv6Address                   = 131
	IPFIX_FIELD_droppedOctetDeltaCount                = 132
	IPFIX_FIELD_droppedPacketDeltaCount               = 133
	IPFIX_FIELD_droppedOctetTotalCount                = 134
	IPFIX_FIELD_droppedPacketTotalCount               = 135
	IPFIX_FIELD_flowEndReason                         = 136
	IPFIX_FIELD_commonPropertiesId                    = 137
	IPFIX_FIELD_observationPointId                    = 138
	IPFIX_FIELD_icmpTypeCodeIPv6                      = 139
	IPFIX_FIELD_mplsTopLabelIPv6Address               = 140
	IPFIX_FIELD_lineCardId                            = 141
	IPFIX_FIELD_portId                                = 142
	IPFIX_FIELD_meteringProcessId                     = 143
	IPFIX_FIELD_exportingProcessId                    = 144
	IPFIX_FIELD_templateId                            = 145
	IPFIX_FIELD_wlanChannelId                         = 146
	IPFIX_FIELD_wlanSSID                              = 147
	IPFIX_FIELD_flowId                                = 148
	IPFIX_FIELD_observationDomainId                   = 149
	IPFIX_FIELD_flowStartSeconds                      = 150
	IPFIX_FIELD_flowEndSeconds                        = 151
	IPFIX_FIELD_flowStartMilliseconds                 = 152
	IPFIX_FIELD_flowEndMilliseconds                   = 153
	IPFIX_FIELD_flowStartMicroseconds                 = 154
	IPFIX_FIELD_flowEndMicroseconds                   = 155
	IPFIX_FIELD_flowStartNanoseconds                  = 156
	IPFIX_FIELD_flowEndNanoseconds                    = 157
	IPFIX_FIELD_flowStartDeltaMicroseconds            = 158
	IPFIX_FIELD_flowEndDeltaMicroseconds              = 159
	IPFIX_FIELD_systemInitTimeMilliseconds            = 160
	IPFIX_FIELD_flowDurationMilliseconds              = 161
	IPFIX_FIELD_flowDurationMicroseconds              = 162
	IPFIX_FIELD_observedFlowTotalCount                = 163
	IPFIX_FIELD_ignoredPacketTotalCount               = 164
	IPFIX_FIELD_ignoredOctetTotalCount                = 165
	IPFIX_FIELD_notSentFlowTotalCount                 = 166
	IPFIX_FIELD_notSentPacketTotalCount               = 167
	IPFIX_FIELD_notSentOctetTotalCount                = 168
	IPFIX_FIELD_destinationIPv6Prefix                 = 169
	IPFIX_FIELD_sourceIPv6Prefix                      = 170
	IPFIX_FIELD_postOctetTotalCount                   = 171
	IPFIX_FIELD_postPacketTotalCount                  = 172
	IPFIX_FIELD_flowKeyIndicator                      = 173
	IPFIX_FIELD_postMCastPacketTotalCount             = 174
	IPFIX_FIELD_postMCastOctetTotalCount              = 175
	IPFIX_FIELD_icmpTypeIPv4                          = 176
	IPFIX_FIELD_icmpCodeIPv4                          = 177
	IPFIX_FIELD_icmpTypeIPv6                          = 178
	IPFIX_FIELD_icmpCodeIPv6                          = 179
	IPFIX_FIELD_udpSourcePort                         = 180
	IPFIX_FIELD_udpDestinationPort                    = 181
	IPFIX_FIELD_tcpSourcePort                         = 182
	IPFIX_FIELD_tcpDestinationPort                    = 183
	IPFIX_FIELD_tcpSequenceNumber                     = 184
	IPFIX_FIELD_tcpAcknowledgementNumber              = 185
	IPFIX_FIELD_tcpWindowSize                         = 186
	IPFIX_FIELD_tcpUrgentPointer                      = 187
	IPFIX_FIELD_tcpHeaderLength                       = 188
	IPFIX_FIELD_ipHeaderLength                        = 189
	IPFIX_FIELD_totalLengthIPv4                       = 190
	IPFIX_FIELD_payloadLengthIPv6                     = 191
	IPFIX_FIELD_ipTTL                                 = 192
	IPFIX_FIELD_nextHeaderIPv6                        = 193
	IPFIX_FIELD_mplsPayloadLength                     = 194
	IPFIX_FIELD_ipDiffServCodePoint                   = 195
	IPFIX_FIELD_ipPrecedence                          = 196
	IPFIX_FIELD_fragmentFlags                         = 197
	IPFIX_FIELD_octetDeltaSumOfSquares                = 198
	IPFIX_FIELD_octetTotalSumOfSquares                = 199
	IPFIX_FIELD_mplsTopLabelTTL                       = 200
	IPFIX_FIELD_mplsLabelStackLength                  = 201
	IPFIX_FIELD_mplsLabelStackDepth                   = 202
	IPFIX_FIELD_mplsTopLabelExp                       = 203
	IPFIX_FIELD_ipPayloadLength                       = 204
	IPFIX_FIELD_udpMessageLength                      = 205
	IPFIX_FIELD_isMulticast                           = 206
	IPFIX_FIELD_ipv4IHL                               = 207
	IPFIX_FIELD_ipv4Options                           = 208
	IPFIX_FIELD_tcpOptions                            = 209
	IPFIX_FIELD_paddingOctets                         = 210
	IPFIX_FIELD_collectorIPv4Address                  = 211
	IPFIX_FIELD_collectorIPv6Address                  = 212
	IPFIX_FIELD_exportInterface                       = 213
	IPFIX_FIELD_exportProtocolVersion                 = 214
	IPFIX_FIELD_exportTransportProtocol               = 215
	IPFIX_FIELD_collectorTransportPort                = 216
	IPFIX_FIELD_exporterTransportPort                 = 217
	IPFIX_FIELD_tcpSynTotalCount                      = 218
	IPFIX_FIELD_tcpFinTotalCount                      = 219
	IPFIX_FIELD_tcpRstTotalCount                      = 220
	IPFIX_FIELD_tcpPshTotalCount                      = 221
	IPFIX_FIELD_tcpAckTotalCount                      = 222
	IPFIX_FIELD_tcpUrgTotalCount                      = 223
	IPFIX_FIELD_ipTotalLength                         = 224
	IPFIX_FIELD_postNATSourceIPv4Address              = 225
	IPFIX_FIELD_postNATDestinationIPv4Address         = 226
	IPFIX_FIELD_postNAPTSourceTransportPort           = 227
	IPFIX_FIELD_postNAPTDestinationTransportPort      = 228
	IPFIX_FIELD_natOriginatingAddressRealm            = 229
	IPFIX_FIELD_natEvent                              = 230
	IPFIX_FIELD_initiatorOctets                       = 231
	IPFIX_FIELD_responderOctets                       = 232
	IPFIX_FIELD_firewallEvent                         = 233
	IPFIX_FIELD_ingressVRFID                          = 234
	IPFIX_FIELD_egressVRFID                           = 235
	IPFIX_FIELD_VRFname                               = 236
	IPFIX_FIELD_postMplsTopLabelExp                   = 237
	IPFIX_FIELD_tcpWindowScale                        = 238
	IPFIX_FIELD_biflowDirection                       = 239
	IPFIX_FIELD_ethernetHeaderLength                  = 240
	IPFIX_FIELD_ethernetPayloadLength                 = 241
	IPFIX_FIELD_ethernetTotalLength                   = 242
	IPFIX_FIELD_dot1qVlanId                           = 243
	IPFIX_FIELD_dot1qPriority                         = 244
	IPFIX_FIELD_dot1qCustomerVlanId                   = 245
	IPFIX_FIELD_dot1qCustomerPriority                 = 246
	IPFIX_FIELD_metroEvcId                            = 247
	IPFIX_FIELD_metroEvcType                          = 248
	IPFIX_FIELD_pseudoWireId                          = 249
	IPFIX_FIELD_pseudoWireType                        = 250
	IPFIX_FIELD_pseudoWireControlWord                 = 251
	IPFIX_FIELD_ingressPhysicalInterface              = 252
	IPFIX_FIELD_egressPhysicalInterface               = 253
	IPFIX_FIELD_postDot1qVlanId                       = 254
	IPFIX_FIELD_postDot1qCustomerVlanId               = 255
	IPFIX_FIELD_ethernetType                          = 256
	IPFIX_FIELD_postIpPrecedence                      = 257
	IPFIX_FIELD_collectionTimeMilliseconds            = 258
	IPFIX_FIELD_exportSctpStreamId                    = 259
	IPFIX_FIELD_maxExportSeconds                      = 260
	IPFIX_FIELD_maxFlowEndSeconds                     = 261
	IPFIX_FIELD_messageMD5Checksum                    = 262
	IPFIX_FIELD_messageScope                          = 263
	IPFIX_FIELD_minExportSeconds                      = 264
	IPFIX_FIELD_minFlowStartSeconds                   = 265
	IPFIX_FIELD_opaqueOctets                          = 266
	IPFIX_FIELD_sessionScope                          = 267
	IPFIX_FIELD_maxFlowEndMicroseconds                = 268
	IPFIX_FIELD_maxFlowEndMilliseconds                = 269
	IPFIX_FIELD_maxFlowEndNanoseconds                 = 270
	IPFIX_FIELD_minFlowStartMicroseconds              = 271
	IPFIX_FIELD_minFlowStartMilliseconds              = 272
	IPFIX_FIELD_minFlowStartNanoseconds               = 273
	IPFIX_FIELD_collectorCertificate                  = 274
	IPFIX_FIELD_exporterCertificate                   = 275
	IPFIX_FIELD_dataRecordsReliability                = 276
	IPFIX_FIELD_observationPointType                  = 277
	IPFIX_FIELD_newConnectionDeltaCount               = 278
	IPFIX_FIELD_connectionSumDurationSeconds          = 279
	IPFIX_FIELD_connectionTransactionId               = 280
	IPFIX_FIELD_postNATSourceIPv6Address              = 281
	IPFIX_FIELD_postNATDestinationIPv6Address         = 282
	IPFIX_FIELD_natPoolId                             = 283
	IPFIX_FIELD_natPoolName                           = 284
	IPFIX_FIELD_anonymizationFlags                    = 285
	IPFIX_FIELD_anonymizationTechnique                = 286
	IPFIX_FIELD_informationElementIndex               = 287
	IPFIX_FIELD_p2pTechnology                         = 288
	IPFIX_FIELD_tunnelTechnology                      = 289
	IPFIX_FIELD_encryptedTechnology                   = 290
	IPFIX_FIELD_basicList                             = 291
	IPFIX_FIELD_subTemplateList                       = 292
	IPFIX_FIELD_subTemplateMultiList                  = 293
	IPFIX_FIELD_bgpValidityState                      = 294
	IPFIX_FIELD_IPSecSPI                              = 295
	IPFIX_FIELD_greKey                                = 296
	IPFIX_FIELD_natType                               = 297
	IPFIX_FIELD_initiatorPackets                      = 298
	IPFIX_FIELD_responderPackets                      = 299
	IPFIX_FIELD_observationDomainName                 = 300
	IPFIX_FIELD_selectionSequenceId                   = 301
	IPFIX_FIELD_selectorId                            = 302
	IPFIX_FIELD_informationElementId                  = 303
	IPFIX_FIELD_selectorAlgorithm                     = 304
	IPFIX_FIELD_samplingPacketInterval                = 305
	IPFIX_FIELD_samplingPacketSpace                   = 306
	IPFIX_FIELD_samplingTimeInterval                  = 307
	IPFIX_FIELD_samplingTimeSpace                     = 308
	IPFIX_FIELD_samplingSize                          = 309
	IPFIX_FIELD_samplingPopulation                    = 310
	IPFIX_FIELD_samplingProbability                   = 311
	IPFIX_FIELD_dataLinkFrameSize                     = 312
	IPFIX_FIELD_ipHeaderPacketSection                 = 313
	IPFIX_FIELD_ipPayloadPacketSection                = 314
	IPFIX_FIELD_dataLinkFrameSection                  = 315
	IPFIX_FIELD_mplsLabelStackSection                 = 316
	IPFIX_FIELD_mplsPayloadPacketSection              = 317
	IPFIX_FIELD_selectorIdTotalPktsObserved           = 318
	IPFIX_FIELD_selectorIdTotalPktsSelected           = 319
	IPFIX_FIELD_absoluteError                         = 320
	IPFIX_FIELD_relativeError                         = 321
	IPFIX_FIELD_observationTimeSeconds                = 322
	IPFIX_FIELD_observationTimeMilliseconds           = 323
	IPFIX_FIELD_observationTimeMicroseconds           = 324
	IPFIX_FIELD_observationTimeNanoseconds            = 325
	IPFIX_FIELD_digestHashValue                       = 326
	IPFIX_FIELD_hashIPPayloadOffset                   = 327
	IPFIX_FIELD_hashIPPayloadSize                     = 328
	IPFIX_FIELD_hashOutputRangeMin                    = 329
	IPFIX_FIELD_hashOutputRangeMax                    = 330
	IPFIX_FIELD_hashSelectedRangeMin                  = 331
	IPFIX_FIELD_hashSelectedRangeMax                  = 332
	IPFIX_FIELD_hashDigestOutput                      = 333
	IPFIX_FIELD_hashInitialiserValue                  = 334
	IPFIX_FIELD_selectorName                          = 335
	IPFIX_FIELD_upperCILimit                          = 336
	IPFIX_FIELD_lowerCILimit                          = 337
	IPFIX_FIELD_confidenceLevel                       = 338
	IPFIX_FIELD_informationElementDataType            = 339
	IPFIX_FIELD_informationElementDescription         = 340
	IPFIX_FIELD_informationElementName                = 341
	IPFIX_FIELD_informationElementRangeBegin          = 342
	IPFIX_FIELD_informationElementRangeEnd            = 343
	IPFIX_FIELD_informationElementSemantics           = 344
	IPFIX_FIELD_informationElementUnits               = 345
	IPFIX_FIELD_privateEnterpriseNumber               = 346
	IPFIX_FIELD_virtualStationInterfaceId             = 347
	IPFIX_FIELD_virtualStationInterfaceName           = 348
	IPFIX_FIELD_virtualStationUUID                    = 349
	IPFIX_FIELD_virtualStationName                    = 350
	IPFIX_FIELD_layer2SegmentId                       = 351
	IPFIX_FIELD_layer2OctetDeltaCount                 = 352
	IPFIX_FIELD_layer2OctetTotalCount                 = 353
	IPFIX_FIELD_ingressUnicastPacketTotalCount        = 354
	IPFIX_FIELD_ingressMulticastPacketTotalCount      = 355
	IPFIX_FIELD_ingressBroadcastPacketTotalCount      = 356
	IPFIX_FIELD_egressUnicastPacketTotalCount         = 357
	IPFIX_FIELD_egressBroadcastPacketTotalCount       = 358
	IPFIX_FIELD_monitoringIntervalStartMilliSeconds   = 359
	IPFIX_FIELD_monitoringIntervalEndMilliSeconds     = 360
	IPFIX_FIELD_portRangeStart                        = 361
	IPFIX_FIELD_portRangeEnd                          = 362
	IPFIX_FIELD_portRangeStepSize                     = 363
	IPFIX_FIELD_portRangeNumPorts                     = 364
	IPFIX_FIELD_staMacAddress                         = 365
	IPFIX_FIELD_staIPv4Address                        = 366
	IPFIX_FIELD_wtpMacAddress                         = 367
	IPFIX_FIELD_ingressInterfaceType                  = 368
	IPFIX_FIELD_egressInterfaceType                   = 369
	IPFIX_FIELD_rtpSequenceNumber                     = 370
	IPFIX_FIELD_userName                              = 371
	IPFIX_FIELD_applicationCategoryName               = 372
	IPFIX_FIELD_applicationSubCategoryName            = 373
	IPFIX_FIELD_applicationGroupName                  = 374
	IPFIX_FIELD_originalFlowsPresent                  = 375
	IPFIX_FIELD_originalFlowsInitiated                = 376
	IPFIX_FIELD_originalFlowsCompleted                = 377
	IPFIX_FIELD_distinctCountOfSourceIPAddress        = 378
	IPFIX_FIELD_distinctCountOfDestinationIPAddress   = 379
	IPFIX_FIELD_distinctCountOfSourceIPv4Address      = 380
	IPFIX_FIELD_distinctCountOfDestinationIPv4Address = 381
	IPFIX_FIELD_distinctCountOfSourceIPv6Address      = 382
	IPFIX_FIELD_distinctCountOfDestinationIPv6Address = 383
	IPFIX_FIELD_valueDistributionMethod               = 384
	IPFIX_FIELD_rfc3550JitterMilliseconds             = 385
	IPFIX_FIELD_rfc3550JitterMicroseconds             = 386
	IPFIX_FIELD_rfc3550JitterNanoseconds              = 387
	IPFIX_FIELD_dot1qDEI                              = 388
	IPFIX_FIELD_dot1qCustomerDEI                      = 389
	IPFIX_FIELD_flowSelectorAlgorithm                 = 390
	IPFIX_FIELD_flowSelectedOctetDeltaCount           = 391
	IPFIX_FIELD_flowSelectedPacketDeltaCount          = 392
	IPFIX_FIELD_flowSelectedFlowDeltaCount            = 393
	IPFIX_FIELD_selectorIDTotalFlowsObserved          = 394
	IPFIX_FIELD_selectorIDTotalFlowsSelected          = 395
	IPFIX_FIELD_samplingFlowInterval                  = 396
	IPFIX_FIELD_samplingFlowSpacing                   = 397
	IPFIX_FIELD_flowSamplingTimeInterval              = 398
	IPFIX_FIELD_flowSamplingTimeSpacing               = 399
	IPFIX_FIELD_hashFlowDomain                        = 400
	IPFIX_FIELD_transportOctetDeltaCount              = 401
	IPFIX_FIELD_transportPacketDeltaCount             = 402
	IPFIX_FIELD_originalExporterIPv4Address           = 403
	IPFIX_FIELD_originalExporterIPv6Address           = 404
	IPFIX_FIELD_originalObservationDomainId           = 405
	IPFIX_FIELD_intermediateProcessId                 = 406
	IPFIX_FIELD_ignoredDataRecordTotalCount           = 407
	IPFIX_FIELD_dataLinkFrameType                     = 408
	IPFIX_FIELD_sectionOffset                         = 409
	IPFIX_FIELD_sectionExportedOctets                 = 410
	IPFIX_FIELD_dot1qServiceInstanceTag               = 411
	IPFIX_FIELD_dot1qServiceInstanceId                = 412
	IPFIX_FIELD_dot1qServiceInstancePriority          = 413
	IPFIX_FIELD_dot1qCustomerSourceMacAddress         = 414
	IPFIX_FIELD_dot1qCustomerDestinationMacAddress    = 415
	IPFIX_FIELD_postLayer2OctetDeltaCount             = 417
	IPFIX_FIELD_postMCastLayer2OctetDeltaCount        = 418
	IPFIX_FIELD_postLayer2OctetTotalCount             = 420
	IPFIX_FIELD_postMCastLayer2OctetTotalCount        = 421
	IPFIX_FIELD_minimumLayer2TotalLength              = 422
	IPFIX_FIELD_maximumLayer2TotalLength              = 423
	IPFIX_FIELD_droppedLayer2OctetDeltaCount          = 424
	IPFIX_FIELD_droppedLayer2OctetTotalCount          = 425
	IPFIX_FIELD_ignoredLayer2OctetTotalCount          = 426
	IPFIX_FIELD_notSentLayer2OctetTotalCount          = 427
	IPFIX_FIELD_layer2OctetDeltaSumOfSquares          = 428
	IPFIX_FIELD_layer2OctetTotalSumOfSquares          = 429
	IPFIX_FIELD_layer2FrameDeltaCount                 = 430
	IPFIX_FIELD_layer2FrameTotalCount                 = 431
	IPFIX_FIELD_pseudoWireDestinationIPv4Address      = 432
	IPFIX_FIELD_ignoredLayer2FrameTotalCount          = 433
	IPFIX_FIELD_mibObjectValueInteger                 = 434
	IPFIX_FIELD_mibObjectValueOctetString             = 435
	IPFIX_FIELD_mibObjectValueOID                     = 436
	IPFIX_FIELD_mibObjectValueBits                    = 437
	IPFIX_FIELD_mibObjectValueIPAddress               = 438
	IPFIX_FIELD_mibObjectValueCounter                 = 439
	IPFIX_FIELD_mibObjectValueGauge                   = 440
	IPFIX_FIELD_mibObjectValueTimeTicks               = 441
	IPFIX_FIELD_mibObjectValueUnsigned                = 442
	IPFIX_FIELD_mibObjectValueTable                   = 443
	IPFIX_FIELD_mibObjectValueRow                     = 444
	IPFIX_FIELD_mibObjectIdentifier                   = 445
	IPFIX_FIELD_mibSubIdentifier                      = 446
	IPFIX_FIELD_mibIndexIndicator                     = 447
	IPFIX_FIELD_mibCaptureTimeSemantics               = 448
	IPFIX_FIELD_mibContextEngineID                    = 449
	IPFIX_FIELD_mibContextName                        = 450
	IPFIX_FIELD_mibObjectName                         = 451
	IPFIX_FIELD_mibObjectDescription                  = 452
	IPFIX_FIELD_mibObjectSyntax                       = 453
	IPFIX_FIELD_mibModuleName                         = 454
	IPFIX_FIELD_mobileIMSI                            = 455
	IPFIX_FIELD_mobileMSISDN                          = 456
	IPFIX_FIELD_httpStatusCode                        = 457
	IPFIX_FIELD_sourceTransportPortsLimit             = 458
	IPFIX_FIELD_httpRequestMethod                     = 459
	IPFIX_FIELD_httpRequestHost                       = 460
	IPFIX_FIELD_httpRequestTarget                     = 461
	IPFIX_FIELD_httpMessageVersion                    = 462
	IPFIX_FIELD_natInstanceID                         = 463
	IPFIX_FIELD_internalAddressRealm                  = 464
	IPFIX_FIELD_externalAddressRealm                  = 465
	IPFIX_FIELD_natQuotaExceededEvent                 = 466
	IPFIX_FIELD_natThresholdEvent                     = 467
)
