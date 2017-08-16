package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type s3Storage struct {
	session  *session.Session
	s3client *s3.S3
	bucket   string
}

// newS3Store initialising an AWS session and then returns
// a new instance of the s3Storage type
func newS3Store(bucket string) s3Storage {
	sess, err := session.NewSession(&aws.Config{Region: aws.String("eu-west-1")})
	if err != nil {
		panic(err)
	}

	svc := s3.New(sess)
	return s3Storage{session: sess, s3client: svc, bucket: bucket}
}

func (s s3Storage) listFiles() (*s3.ListObjectsOutput, error) {
	input := &s3.ListObjectsInput{
		Bucket:  aws.String(s.bucket),
		MaxKeys: aws.Int64(5),
	}

	result, err := s.s3client.ListObjects(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchBucket:
				fmt.Println(s3.ErrCodeNoSuchBucket, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return nil, err
	}

	return result, nil
}

func (s s3Storage) getFile(key string) (*s3.GetObjectOutput, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	result, err := s.s3client.GetObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				fmt.Println(s3.ErrCodeNoSuchKey, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return nil, err
	}

	return result, nil
}

func (s s3Storage) storeFile(name string, content []byte) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(name),
		Body:   bytes.NewReader(content),
	}

	_, err := s.s3client.PutObject(input)
	return err
}

func (s s3Storage) getPage(p *wikiPage) (*wikiPage, error) {
	filename := getWikiFilename(wikiDir, p.Title)
	file, err := s.getFile(filename)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(file.Body)
	if err != nil {
		return nil, err
	}
	p.Body = template.HTML(body)
	p.Modified = file.LastModified.Format("2006-01-02 15:04:05")
	_, someTags := file.Metadata["tags"]
	if someTags {
		p.Tags = *file.Metadata["tags"]
	}
	if len(p.Tags) > 0 {
		p.TagArray = strings.Split(p.Tags, ",")
	}

	_, pubFlag := file.Metadata["published"]
	if pubFlag {
		pub, err := strconv.ParseBool(*file.Metadata["published"])
		if err == nil {
			p.Published = pub
		}
	}

	return p, nil
}
