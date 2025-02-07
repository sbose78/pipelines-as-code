package kubernetestint

import (
	"context"
	"errors"
)

type KinterfaceTest struct {
	ConsoleURL     string
	NamespaceError bool

	prDescribe string
}

func (k *KinterfaceTest) GetConsoleUI(ctx context.Context, ns string, pr string) (string, error) {
	return k.ConsoleURL, nil
}

func (k *KinterfaceTest) GetNamespace(ctx context.Context, ns string) error {
	if k.NamespaceError {
		return errors.New("cannot find Namespace")
	}
	return nil
}

func (k *KinterfaceTest) TektonCliPRDescribe(namespace, prName string) (string, error) {
	return k.prDescribe, nil
}

func (k *KinterfaceTest) TektonCliFollowLogs(prName, namespace string) (string, error) {
	return k.prDescribe, nil
}
