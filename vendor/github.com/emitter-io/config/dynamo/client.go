// Copyright (c) Roman Atachiants and contributors. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for details.

package dynamo

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

var errKeyNotFound = errors.New("key was not found in dynamodb")

// client represents a dynamodb client for secret storage
type client struct {
	dynamo    dynamodbiface.DynamoDBAPI
	table     string
	keyColumn string
	valColumn string
}

// newClient creates a new client.
func newClient(region, table, keyColumn, valColumn string) (*client, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)

	if err != nil {
		return nil, err
	}

	return &client{
		dynamo:    dynamodb.New(sess),
		table:     table,
		keyColumn: keyColumn,
		valColumn: valColumn,
	}, nil
}

// Put stores the data
func (c *client) Put(key string, data string) error {
	_, err := c.dynamo.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(c.table),
		Item: map[string]*dynamodb.AttributeValue{
			c.keyColumn: {S: aws.String(key)},
			c.valColumn: {S: aws.String(data)},
		},
	})
	return err
}

// Get returns a value.
func (c *client) Get(key string) (string, error) {
	result, err := c.dynamo.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(c.table),
		Key:       c.newKey(key),
	})

	if err != nil {
		return "", err
	}

	// Get the value column, must be binary
	if v, ok := result.Item[c.valColumn]; ok && v.S != nil {
		return *v.S, nil
	}

	return "", errKeyNotFound
}

// Delete removes a value.
func (c *client) Delete(key string) error {
	_, err := c.dynamo.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String(c.table),
		Key:       c.newKey(key),
	})
	return err
}

// NewKey prepare a key
func (c *client) newKey(key string) map[string]*dynamodb.AttributeValue {
	return map[string]*dynamodb.AttributeValue{
		c.keyColumn: {S: aws.String(key)},
	}
}
