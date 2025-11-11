package dynamicconfig

import (
	"fmt"
	"strings"
	"sync/atomic"
)

type (
	registry struct {
		settings map[string]GenericSetting
		queried  atomic.Bool
	}
)

var (
	globalRegistry registry
)

func register(s GenericSetting) {
	if globalRegistry.queried.Load() {
		panic("dynamicconfig.New*Setting must only be called from static initializers")
	}
	if globalRegistry.settings == nil {
		globalRegistry.settings = make(map[string]GenericSetting)
	}
	keyStr := strings.ToLower(s.Key().String())
	if globalRegistry.settings[keyStr] != nil {
		panic(fmt.Sprintf("duplicate registration of dynamic config key: %q", keyStr))
	}
	globalRegistry.settings[keyStr] = s
}

func queryRegistry(k Key) GenericSetting {
	if !globalRegistry.queried.Load() {
		globalRegistry.queried.Store(true)
	}
	return globalRegistry.settings[strings.ToLower(k.String())]
}

func GetAllFrontendSettingsFromRegistry() map[string]GenericSetting {
	if !globalRegistry.queried.Load() {
		globalRegistry.queried.Store(true)
	}

	return globalRegistry.settings
	// result := make(map[string]GenericSetting, len(globalRegistry.settings))
	// for k, v := range globalRegistry.settings {
	// 	result[k] = v
	// }
	// return result
}

// For testing only; do not call from regular code!
func ResetRegistryForTest() {
	globalRegistry.settings = nil
	globalRegistry.queried.Store(false)
}

func FrontendMapRegistry() *map[string]interface{} {
	return &map[string]interface{}{
		"PersistenceMaxQPS":                                                 FrontendPersistenceMaxQPS,
		"PersistenceGlobalMaxQPS":                                           FrontendPersistenceGlobalMaxQPS,
		"PersistenceNamespaceMaxQPS":                                        FrontendPersistenceNamespaceMaxQPS,
		"PersistenceGlobalNamespaceMaxQPS":                                  FrontendPersistenceGlobalNamespaceMaxQPS,
		"PersistencePerShardNamespaceMaxQPS":                                DefaultPerShardNamespaceRPSMax,
		"PersistenceDynamicRateLimitingParams":                              FrontendPersistenceDynamicRateLimitingParams,
		"PersistenceQPSBurstRatio":                                          PersistenceQPSBurstRatio,
		"VisibilityPersistenceMaxReadQPS":                                   VisibilityPersistenceMaxReadQPS,
		"VisibilityPersistenceMaxWriteQPS":                                  VisibilityPersistenceMaxWriteQPS,
		"VisibilityPersistenceSlowQueryThreshold":                           VisibilityPersistenceSlowQueryThreshold,
		"VisibilityMaxPageSize":                                             FrontendVisibilityMaxPageSize,
		"EnableReadFromSecondaryVisibility":                                 EnableReadFromSecondaryVisibility,
		"VisibilityEnableShadowReadMode":                                    VisibilityEnableShadowReadMode,
		"VisibilityDisableOrderByClause":                                    VisibilityDisableOrderByClause,
		"VisibilityEnableManualPagination":                                  VisibilityEnableManualPagination,
		"VisibilityAllowList":                                               VisibilityAllowList,
		"SuppressErrorSetSystemSearchAttribute":                             SuppressErrorSetSystemSearchAttribute,
		"HistoryMaxPageSize":                                                FrontendHistoryMaxPageSize,
		"RPS":                                                               FrontendRPS,
		"GlobalRPS":                                                         FrontendGlobalRPS,
		"OperatorRPSRatio":                                                  OperatorRPSRatio,
		"NamespaceReplicationInducingAPIsRPS":                               FrontendNamespaceReplicationInducingAPIsRPS,
		"MaxNamespaceRPSPerInstance":                                        FrontendMaxNamespaceRPSPerInstance,
		"MaxNamespaceBurstRatioPerInstance":                                 FrontendMaxNamespaceBurstRatioPerInstance,
		"MaxConcurrentLongRunningRequestsPerInstance":                       FrontendMaxConcurrentLongRunningRequestsPerInstance,
		"MaxGlobalConcurrentLongRunningRequests":                            FrontendGlobalMaxConcurrentLongRunningRequests,
		"MaxNamespaceVisibilityRPSPerInstance":                              FrontendMaxNamespaceVisibilityRPSPerInstance,
		"MaxNamespaceVisibilityBurstRatioPerInstance":                       FrontendMaxNamespaceVisibilityBurstRatioPerInstance,
		"MaxNamespaceNamespaceReplicationInducingAPIsRPSPerInstance":        FrontendMaxNamespaceNamespaceReplicationInducingAPIsRPSPerInstance,
		"MaxNamespaceNamespaceReplicationInducingAPIsBurstRatioPerInstance": FrontendMaxNamespaceNamespaceReplicationInducingAPIsBurstRatioPerInstance,
		"GlobalNamespaceRPS":                                                FrontendGlobalNamespaceRPS,
		"InternalFEGlobalNamespaceRPS":                                      InternalFrontendGlobalNamespaceRPS,
		"GlobalNamespaceVisibilityRPS":                                      FrontendGlobalNamespaceVisibilityRPS,
		"InternalFEGlobalNamespaceVisibilityRPS":                            InternalFrontendGlobalNamespaceVisibilityRPS,
		"GlobalNamespaceNamespaceReplicationInducingAPIsRPS":                FrontendGlobalNamespaceNamespaceReplicationInducingAPIsRPS,
		"MaxIDLengthLimit":                                                  MaxIDLengthLimit,
		"WorkerBuildIdSizeLimit":                                            WorkerBuildIdSizeLimit,
		"ReachabilityTaskQueueScanLimit":                                    ReachabilityTaskQueueScanLimit,
		"ReachabilityQueryBuildIdLimit":                                     ReachabilityQueryBuildIdLimit,
		"ReachabilityCacheOpenWFsTTL":                                       ReachabilityCacheOpenWFsTTL,
		"ReachabilityCacheClosedWFsTTL":                                     ReachabilityCacheClosedWFsTTL,
		"ReachabilityQuerySetDurationSinceDefault":                          ReachabilityQuerySetDurationSinceDefault,
		"MaxBadBinaries":                                                    FrontendMaxBadBinaries,
		"DisableListVisibilityByFilter":                                     DisableListVisibilityByFilter,
		"BlobSizeLimitError":                                                BlobSizeLimitError,
		"BlobSizeLimitWarn":                                                 BlobSizeLimitWarn,
		"ThrottledLogRPS":                                                   FrontendThrottledLogRPS,
		"ShutdownDrainDuration":                                             FrontendShutdownDrainDuration,
		"ShutdownFailHealthCheckDuration":                                   FrontendShutdownFailHealthCheckDuration,
		"EnableNamespaceNotActiveAutoForwarding":                            EnableNamespaceNotActiveAutoForwarding,
		"SearchAttributesNumberOfKeysLimit":                                 SearchAttributesNumberOfKeysLimit,
		"SearchAttributesSizeOfValueLimit":                                  SearchAttributesSizeOfValueLimit,
		"SearchAttributesTotalSizeLimit":                                    SearchAttributesTotalSizeLimit,
		"VisibilityArchivalQueryMaxPageSize":                                VisibilityArchivalQueryMaxPageSize,
		"DisallowQuery":                                                     DisallowQuery,
		"SendRawWorkflowHistory":                                            SendRawWorkflowHistory,
		"DefaultWorkflowRetryPolicy":                                        DefaultWorkflowRetryPolicy,
		"DefaultWorkflowTaskTimeout":                                        DefaultWorkflowTaskTimeout,
		"EnableServerVersionCheck":                                          EnableServerVersionCheck,
		"EnableTokenNamespaceEnforcement":                                   EnableTokenNamespaceEnforcement,
		"ExposeAuthorizerErrors":                                            ExposeAuthorizerErrors,
		"KeepAliveMinTime":                                                  KeepAliveMinTime,
		"KeepAlivePermitWithoutStream":                                      KeepAlivePermitWithoutStream,
		"KeepAliveMaxConnectionIdle":                                        KeepAliveMaxConnectionIdle,
		"KeepAliveMaxConnectionAge":                                         KeepAliveMaxConnectionAge,
		"KeepAliveMaxConnectionAgeGrace":                                    KeepAliveMaxConnectionAgeGrace,
		"KeepAliveTime":                                                     KeepAliveTime,
		"KeepAliveTimeout":                                                  KeepAliveTimeout,
		"DeleteNamespaceDeleteActivityRPS":                                  DeleteNamespaceDeleteActivityRPS,
		"DeleteNamespacePageSize":                                           DeleteNamespacePageSize,
		"DeleteNamespacePagesPerExecution":                                  DeleteNamespacePagesPerExecution,
		"DeleteNamespaceConcurrentDeleteExecutionsActivities":               DeleteNamespaceConcurrentDeleteExecutionsActivities,
		"DeleteNamespaceNamespaceDeleteDelay":                               DeleteNamespaceNamespaceDeleteDelay,
		"EnableSchedules":                                                   FrontendEnableSchedules,
		"EnableDeployments":                                                 EnableDeployments,
		"EnableDeploymentVersions":                                          EnableDeploymentVersions,
		"EnableBatcher":                                                     FrontendEnableBatcher,
		"MaxConcurrentBatchOperation":                                       FrontendMaxConcurrentBatchOperationPerNamespace,
		"MaxExecutionCountBatchOperation":                                   FrontendMaxExecutionCountBatchOperationPerNamespace,
		"EnableExecuteMultiOperation":                                       FrontendEnableExecuteMultiOperation,
		"EnableUpdateWorkflowExecution":                                     FrontendEnableUpdateWorkflowExecution,
		"EnableUpdateWorkflowExecutionAsyncAccepted":                        FrontendEnableUpdateWorkflowExecutionAsyncAccepted,
		"NumConsecutiveWorkflowTaskProblemsToTriggerSearchAttribute":        NumConsecutiveWorkflowTaskProblemsToTriggerSearchAttribute,
		"EnableWorkerVersioningData":                                        FrontendEnableWorkerVersioningDataAPIs,
		"EnableWorkerVersioningWorkflow":                                    FrontendEnableWorkerVersioningWorkflowAPIs,
		"EnableWorkerVersioningRules":                                       FrontendEnableWorkerVersioningRuleAPIs,
		"EnableNexusAPIs":                                                   EnableNexus,
		"CallbackURLMaxLength":                                              FrontendCallbackURLMaxLength,
		"CallbackHeaderMaxSize":                                             FrontendCallbackHeaderMaxSize,
		"MaxCallbacksPerWorkflow":                                           MaxCallbacksPerWorkflow,
		// "MaxNexusOperationTokenLength":                                      MaxOperationTokenLength,
		"NexusRequestHeadersBlacklist":   FrontendNexusRequestHeadersBlacklist,
		"NexusForwardRequestUseEndpoint": FrontendNexusForwardRequestUseEndpointDispatch,
		// "NexusOperationsMetricTagConfig":                                    MetricTagConfiguration,
		"LinkMaxSize":        FrontendLinkMaxSize,
		"MaxLinksPerRequest": FrontendMaxLinksPerRequest,
		// "CallbackEndpointConfigs":                                           AllowedAddresses,
		"AdminEnableListHistoryTasks":    AdminEnableListHistoryTasks,
		"MaskInternalErrorDetails":       FrontendMaskInternalErrorDetails,
		"HistoryHostErrorPercentage":     HistoryHostErrorPercentage,
		"HistoryHostSelfErrorProportion": HistoryHostSelfErrorProportion,
		"LogAllReqErrors":                LogAllReqErrors,
		"EnableEagerWorkflowStart":       EnableEagerWorkflowStart,
		"WorkflowRulesAPIsEnabled":       WorkflowRulesAPIsEnabled,
		"MaxWorkflowRulesPerNamespace":   MaxWorkflowRulesPerNamespace,
		"WorkerHeartbeatsEnabled":        WorkerHeartbeatsEnabled,
		"ListWorkersEnabled":             ListWorkersEnabled,
		"WorkerCommandsEnabled":          WorkerCommandsEnabled,
		"HTTPAllowedHosts":               FrontendHTTPAllowedHosts,
	}
}
