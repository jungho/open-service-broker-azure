package servicebus

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/open-service-broker-azure/pkg/service"
)

func (s *serviceManager) GetDeprovisioner(
	service.Plan,
) (service.Deprovisioner, error) {
	return service.NewDeprovisioner(
		service.NewDeprovisioningStep("deleteARMDeployment", s.deleteARMDeployment),
		service.NewDeprovisioningStep("deleteNamespace", s.deleteNamespace),
	)
}

func (s *serviceManager) deleteARMDeployment(
	_ context.Context,
	instance service.Instance,
) (service.InstanceDetails, error) {
	dt, ok := instance.Details.(*serviceBusInstanceDetails)
	if !ok {
		return nil, fmt.Errorf(
			"error casting instance.Details as *serviceBusInstanceDetails",
		)
	}
	if err := s.armDeployer.Delete(
		dt.ARMDeploymentName,
		instance.ResourceGroup,
	); err != nil {
		return nil, fmt.Errorf("error deleting ARM deployment: %s", err)
	}
	return dt, nil
}

func (s *serviceManager) deleteNamespace(
	ctx context.Context,
	instance service.Instance,
) (service.InstanceDetails, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	dt, ok := instance.Details.(*serviceBusInstanceDetails)
	if !ok {
		return nil, fmt.Errorf(
			"error casting instance.Details as *serviceBusInstanceDetails",
		)
	}
	result, err := s.namespacesClient.Delete(
		ctx,
		instance.ResourceGroup,
		dt.ServiceBusNamespaceName,
	)
	if err != nil {
		return nil, fmt.Errorf("error deleting service bus namespace: %s", err)
	}
	if err := result.WaitForCompletion(
		ctx,
		s.namespacesClient.Client,
	); err != nil {
		// Workaround for https://github.com/Azure/azure-sdk-for-go/issues/759
		if strings.Contains(err.Error(), "StatusCode=404") {
			return dt, nil
		}
		return nil, fmt.Errorf("error deleting service bus namespace: %s", err)
	}
	return dt, nil
}
