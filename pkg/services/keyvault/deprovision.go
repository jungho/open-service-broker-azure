package keyvault

import (
	"context"
	"fmt"

	"github.com/Azure/open-service-broker-azure/pkg/service"
)

func (s *serviceManager) GetDeprovisioner(
	service.Plan,
) (service.Deprovisioner, error) {
	return service.NewDeprovisioner(
		service.NewDeprovisioningStep("deleteARMDeployment", s.deleteARMDeployment),
		service.NewDeprovisioningStep(
			"deleteKeyVaultServer",
			s.deleteKeyVaultServer,
		),
	)
}

func (s *serviceManager) deleteARMDeployment(
	_ context.Context,
	instance service.Instance,
) (service.InstanceDetails, error) {
	dt, ok := instance.Details.(*keyvaultInstanceDetails)
	if !ok {
		return nil, fmt.Errorf(
			"error casting instance.Details as *keyvaultInstanceDetails",
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

func (s *serviceManager) deleteKeyVaultServer(
	ctx context.Context,
	instance service.Instance,
) (service.InstanceDetails, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	dt, ok := instance.Details.(*keyvaultInstanceDetails)
	if !ok {
		return nil, fmt.Errorf(
			"error casting instance.Details as *keyvaultInstanceDetails",
		)
	}
	if _, err := s.vaultsClient.Delete(
		ctx,
		instance.ResourceGroup,
		dt.KeyVaultName,
	); err != nil {
		return nil, fmt.Errorf("error deleting keyvault: %s", err)
	}
	return dt, nil
}
