package storage

import (
	"io"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// [
// 	"mediaDisk" : {
// 		Bucket: "",
// 		Region: "",
// 		Profile: "",
// 	},
// 	"pdfDisk" : {
// 		Bucket: "",
// 		Region: "",
// 		Profile: "",
// 	},
// ]

type DiskStruct struct {
	Bucket string
	Region string
	Profile string
}

type StorageStruct struct {
	Disks map[string]DiskStruct
	connection *s3.S3
	CurrentDisk string
}

func InitStorage(currentDisk string, disks map[string]DiskStruct) *StorageStruct {
	storage := &StorageStruct{
		CurrentDisk: currentDisk,
		Disks:    disks,
	}
	storage.SetCurrentDisk(currentDisk) 
	return storage
}

func (st *StorageStruct) SetCurrentDisk(diskName string) {
	if _, exists := st.Disks[diskName]; exists {
		st.CurrentDisk = diskName
		st.connection = nil
	} else {
		panic("No Disk found")
	}
}

func (st *StorageStruct) InitConnection() {
	if st.connection == nil {
		disk := st.Disks[st.CurrentDisk]
		sess, err := session.NewSession(&aws.Config{
			Region:      aws.String(disk.Region),
			Credentials: credentials.NewSharedCredentials("", disk.Profile),
		})
		if err != nil {
			panic("Unable to create AWS session: " + err.Error())
		}
		st.connection = s3.New(sess)
	}
}

func (st *StorageStruct) GetBucketItems() (s3.ListObjectsV2Output, error) {
	st.InitConnection()
	currentDiskInfo := st.getDiskInfo()

	resp, err := st.connection.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(currentDiskInfo.Bucket)})
	if err != nil {
		panic(err)
	}
	return *resp, nil
}

func (st *StorageStruct) Put(path string, body io.Reader) (*s3manager.UploadOutput, error) {
	st.InitConnection()
	currentDiskInfo := st.getDiskInfo()

	sess, err := session.NewSession(&aws.Config{
    	Region: aws.String(currentDiskInfo.Region),
		Credentials: credentials.NewSharedCredentials("", currentDiskInfo.Profile),
	})
	if err != nil {
		return nil, err
	}
	uploader := s3manager.NewUploader(sess)

	output, err := uploader.Upload(&s3manager.UploadInput{
    	Bucket: aws.String(currentDiskInfo.Bucket),
    	Key: aws.String(path),
    	Body: body,
	})
	if err != nil {
		return nil, err
	}
	
	return output, nil
}

func (st *StorageStruct) Delete(path string) (bool, error) {
	st.InitConnection()
	currentDiskInfo := st.getDiskInfo()
	
	_, err := st.connection.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(currentDiskInfo.Bucket), Key: aws.String(path)})

	if err != nil {
		return false, err
	}

	watchErr := st.connection.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(currentDiskInfo.Bucket),
		Key:    aws.String(path),
	})

	if watchErr != nil {
		return false, watchErr
	}

	return true, nil
}

func (st *StorageStruct) Exists(path string) (s3.HeadObjectOutput, error) {
	st.InitConnection()
	currentDiskInfo := st.getDiskInfo()

	res, err := st.GetObjectHead(currentDiskInfo.Bucket, path)
	if err != nil {
		panic(err)
	}
	return res,nil
}

func (st *StorageStruct) SignedURL(path string, expiry int) (map[string]string, error) {
	st.InitConnection()
	currentDiskInfo := st.getDiskInfo()

	putReq, _ := st.connection.PutObjectRequest(&s3.PutObjectInput{
        Bucket: aws.String(currentDiskInfo.Bucket),
        Key:    aws.String(path),
    })
    putUrl, err := putReq.Presign(time.Duration(expiry) * time.Minute)
	if err != nil {
		return map[string]string{}, err
	}

	getReq, _ := st.connection.GetObjectRequest(&s3.GetObjectInput{
        Bucket: aws.String(currentDiskInfo.Bucket),
        Key:    aws.String(path),
    })
    getUrl, err := getReq.Presign(time.Duration(expiry) * time.Minute)
	if err != nil {
		return map[string]string{}, err
	}

	return map[string]string{"signedUrl": putUrl, "url": getUrl}, err
}

func (st *StorageStruct) Meta(path string) (s3.HeadObjectOutput, error) {
	st.InitConnection()
	currentDiskInfo := st.getDiskInfo()

	res, err := st.GetObjectHead(currentDiskInfo.Bucket, path)
	if err != nil {
		panic(err)
	}
	return res,nil
}

func (st *StorageStruct) GetObjectHead(bucket string, path string) (s3.HeadObjectOutput, error) {
	st.InitConnection()
	output, err := st.connection.HeadObject(&s3.HeadObjectInput{
		Bucket: &bucket,
		Key: &path,
	})
	if err != nil {
		panic(err)
	}
	return *output, nil
}

func (st *StorageStruct) getDiskInfo() DiskStruct{
	return st.Disks[st.CurrentDisk]
}