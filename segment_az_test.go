package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type azArgs struct {
	enabled          bool
	subscriptionName string
	subscriptionID   string
	infoSeparator    string
	displayID        bool
	displayName      bool
}

func bootStrapAzTest(args *azArgs) *az {
	env := new(MockedEnvironment)
	env.On("hasCommand", "az").Return(args.enabled)
	env.On("runCommand", "az", []string{"account", "show", "--query=[name,id]", "-o=tsv"}).Return(fmt.Sprintf("%s\n%s\n", args.subscriptionName, args.subscriptionID), nil)
	props := &properties{
		values: map[Property]interface{}{
			SubscriptionInfoSeparator: args.infoSeparator,
			DisplaySubscriptionID:     args.displayID,
			DisplaySubscriptionName:   args.displayName,
		},
	}

	a := &az{
		env:   env,
		props: props,
	}
	return a
}

func TestEnabledAzNotFound(t *testing.T) {
	args := &azArgs{
		enabled: false,
	}
	az := bootStrapAzTest(args)
	assert.False(t, az.enabled())
}

func TestEnabledNoAzDataToDisplay(t *testing.T) {
	args := &azArgs{
		enabled:     true,
		displayID:   false,
		displayName: false,
	}
	az := bootStrapAzTest(args)
	assert.False(t, az.enabled())
}

func TestWriteAzSubscriptionId(t *testing.T) {
	expected := "id"
	args := &azArgs{
		enabled:          true,
		subscriptionID:   "id",
		subscriptionName: "name",
		displayID:        true,
		displayName:      false,
	}
	az := bootStrapAzTest(args)
	assert.True(t, az.enabled())
	assert.Equal(t, expected, az.string())
}

func TestWriteAzSubscriptionName(t *testing.T) {
	expected := "name"
	args := &azArgs{
		enabled:          true,
		subscriptionID:   "id",
		subscriptionName: "name",
		displayID:        false,
		displayName:      true,
	}
	az := bootStrapAzTest(args)
	assert.True(t, az.enabled())
	assert.Equal(t, expected, az.string())
}

func TestWriteAzNameAndID(t *testing.T) {
	expected := "name@id"
	args := &azArgs{
		enabled:          true,
		subscriptionID:   "id",
		subscriptionName: "name",
		infoSeparator:    "@",
		displayID:        true,
		displayName:      true,
	}
	az := bootStrapAzTest(args)
	assert.True(t, az.enabled())
	assert.Equal(t, expected, az.string())
}
