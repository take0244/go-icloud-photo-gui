package ificloud

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"

	"github.com/take0244/go-icloud-photo-gui/aop"
	"github.com/take0244/go-icloud-photo-gui/util"
)

type (
	photoService struct {
		webServiceSckdatabasewsUrl string
		httpClient                 *http.Client
	}
	Photo struct {
		responseRecord
	}
	response struct {
		Records []responseRecord `json:"records"`
	}
	responseRecord struct {
		RecordType string         `json:"recordType"`
		Fields     map[string]any `json:"fields"`
		RecordName string         `json:"recordName"`
	}
)

func (p *photoService) GetAllPhotos() []Photo {
	aop.Logger().Debug("photoService.GetAllPhotos")
	var photos []Photo
	offset := int64(0)
	listType := "CPLAssetAndMasterByAddedDate"
	direction := "ASCENDING"
	url := util.MustParseUrl(
		p.webServiceSckdatabasewsUrl+"/database/1/com.apple.photos.cloud/production/private/records/query",
		map[string]string{
			"remapEnums":          "True",
			"getCurrentSyncToken": "True",
		},
	)
	headers := map[string]string{
		"Content-Type":    "text/plain",
		"Accept-Encoding": "gzip, deflate",
		"Accept":          "*/*",
		"Connection":      "keep-alive",
		"Origin":          "https://www.icloud.com",
		"Referer":         "https://www.icloud.com/",
		"User-Agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36",
	}

	for {
		req := util.MustRequest(
			http.MethodPost,
			url,
			bytes.NewBuffer(util.MustMarshal(map[string]any{
				"query": map[string]any{
					"filterBy": []map[string]any{
						{
							"fieldName":  "startRank",
							"fieldValue": map[string]any{"type": "INT64", "value": offset},
							"comparator": "EQUALS",
						},
						{
							"fieldName":  "direction",
							"fieldValue": map[string]any{"type": "STRING", "value": direction},
							"comparator": "EQUALS",
						},
					},
					"recordType": listType,
				},
				"resultsLimit": 200,
				"desiredKeys": []string{
					"resJPEGFullWidth", "resJPEGFullHeight", "resJPEGFullFileType", "resJPEGFullFingerprint", "resJPEGFullRes",
					"resJPEGLargeWidth", "resJPEGLargeHeight", "resJPEGLargeFileType", "resJPEGLargeFingerprint", "resJPEGLargeRes",
					"resJPEGMedWidth", "resJPEGMedHeight", "resJPEGMedFileType", "resJPEGMedFingerprint", "resJPEGMedRes",
					"resJPEGThumbWidth", "resJPEGThumbHeight", "resJPEGThumbFileType", "resJPEGThumbFingerprint", "resJPEGThumbRes",
					"resVidFullWidth", "resVidFullHeight", "resVidFullFileType", "resVidFullFingerprint", "resVidFullRes",
					"resVidMedWidth", "resVidMedHeight", "resVidMedFileType", "resVidMedFingerprint", "resVidMedRes",
					"resVidSmallWidth", "resVidSmallHeight", "resVidSmallFileType", "resVidSmallFingerprint", "resVidSmallRes",
					"resSidecarWidth", "resSidecarHeight", "resSidecarFileType", "resSidecarFingerprint", "resSidecarRes",
					"itemType", "dataClassType", "filenameEnc", "originalOrientation", "resOriginalWidth", "resOriginalHeight",
					"resOriginalFileType", "resOriginalFingerprint", "resOriginalRes", "resOriginalAltWidth", "resOriginalAltHeight",
					"resOriginalAltFileType", "resOriginalAltFingerprint", "resOriginalAltRes", "resOriginalVidComplWidth",
					"resOriginalVidComplHeight", "resOriginalVidComplFileType", "resOriginalVidComplFingerprint", "resOriginalVidComplRes",
					"isDeleted", "isExpunged", "dateExpunged", "remappedRef", "recordName", "recordType", "recordChangeTag",
					"masterRef", "adjustmentRenderType", "assetDate", "addedDate", "isFavorite", "isHidden", "orientation", "duration",
					"assetSubtype", "assetSubtypeV2", "assetHDRType", "burstFlags", "burstFlagsExt", "burstId", "captionEnc",
					"locationEnc", "locationV2Enc", "locationLatitude", "locationLongitude", "adjustmentType", "timeZoneOffset",
					"vidComplDurValue", "vidComplDurScale", "vidComplDispValue", "vidComplDispScale", "vidComplVisibilityState",
					"customRenderedValue", "containerId", "itemId", "position", "isKeyAsset",
				},
				"zoneID": map[string]string{"zoneName": "PrimarySync"},
			})),
			headers,
		)

		resp, err := util.HttpDoGzipJSON[response](p.httpClient, req)
		if err != nil {
			return nil
		}

		assetRecords := map[string]responseRecord{}
		masterRecords := []responseRecord{}

		for _, rec := range resp.Records {
			if rec.RecordType == "CPLAsset" {
				masterID := rec.Fields["masterRef"].(map[string]any)["value"].(map[string]any)["recordName"].(string)
				assetRecords[masterID] = rec
			} else if rec.RecordType == "CPLMaster" {
				masterRecords = append(masterRecords, rec)
			}
		}

		for _, masterRecord := range masterRecords {
			if asset, exists := assetRecords[masterRecord.RecordName]; exists {
				photos = append(photos, Photo{asset})
			}
		}

		if len(masterRecords) == 0 {
			break
		}

		offset += int64(len(masterRecords))
	}

	return photos
}

func (p *photoService) DownloadAllPhotos(dir string) error {
	aop.Logger().Debug("photoService.DownloadAllPhotos")
	photos := p.GetAllPhotos()
	var allPhotoNames []string
	for _, photo := range photos {
		allPhotoNames = append(allPhotoNames, photo.RecordName)
	}

	url := util.MustParseUrl(
		p.webServiceSckdatabasewsUrl+"/database/1/com.apple.photos.cloud/production/private/records/zip/prepare",
		map[string]string{
			"remapEnums":          "True",
			"getCurrentSyncToken": "True",
		},
	)

	downloadUrls := []string{}
	includeRecordsList := util.ChunkSlice(allPhotoNames, 1000)
	for idx, includeRecords := range includeRecordsList {
		body := bytes.NewBuffer(util.MustMarshal(map[string]any{
			"includeRecords": includeRecords,
			"archiveName":    "iCloud_" + strconv.Itoa(idx) + ".zip",
			"zoneID": map[string]any{
				"zoneName":        "PrimarySync",
				"ownerRecordName": "_2bafdbd84302e4f4dc838ed4f58d41dd",
				"zoneType":        "REGULAR_CUSTOM_ZONE",
			},
			"pluginFields": map[string]any{
				"originalsOnly": map[string]any{
					"value": 1,
					"type":  "INT64",
				},
				"codecs": map[string]any{
					"value": []string{"HEVC", "H.264"},
					"type":  "STRING_LIST",
				},
				"itemTypes": map[string]any{
					"value": []string{"public.heic", "public.jpeg", "public.png", "com.compuserve.gif", "com.apple.m4v-video", "com.apple.quicktime-movie", "public.mpeg-4"},
					"type":  "STRING_LIST",
				},
			},
		}))
		headers := map[string]string{
			"Content-Type": "text/plain;charset=UTF-8",
			"Accept":       "*/*",
			"Connection":   "keep-alive",
			"Origin":       "https://www.icloud.com",
			"Referer":      "https://www.icloud.com/",
			"User-Agent":   "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36",
		}

		req := util.MustRequest(http.MethodPost, url, body, headers)
		resp, err := util.HttpDoJSON[map[string]any](p.httpClient, req)
		if err != nil {
			return fmt.Errorf("failed to zip request: %w", err)
		}

		downloadUrls = append(downloadUrls, (*resp)["downloadURL"].(string))
	}

	for _, url := range downloadUrls {
		if err := util.HttpDownloadAndUnzip(url, dir); err != nil {
			return err
		}
	}

	return nil
}
