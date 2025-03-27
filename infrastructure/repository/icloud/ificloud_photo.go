package infraicloud

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/take0244/go-icloud-photo-gui/aop"
	"github.com/take0244/go-icloud-photo-gui/appctx"
	"github.com/take0244/go-icloud-photo-gui/usecase"
	"github.com/take0244/go-icloud-photo-gui/util"
)

type (
	photoService struct{}
	Photo        struct {
		Fields       map[string]any
		MasterFields map[string]any
		RecordName   string
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

func (p *photoService) getPhotos(ctx context.Context, offset int64) ([]Photo, error) {
	appctx.AppTrace(ctx)
	defer appctx.DeferAppTrace(ctx)

	httpClient, _, appleInfo, _ := MetaData(ctx)

	if aop.IsDebug() {
		ctx = util.WithCache(ctx, true)
	}

	url := util.MustParseUrl(
		appleInfo.WebServiceSckdatabasewsUrl+"/database/1/com.apple.photos.cloud/production/private/records/query",
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
		"User-Agent":      util.UserAgent,
	}

	resp, err := util.HttpDoGzipJSON[response](
		httpClient,
		util.MustRequest(
			ctx,
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
							"fieldValue": map[string]any{"type": "STRING", "value": "ASCENDING"},
							"comparator": "EQUALS",
						},
					},
					"recordType": "CPLAssetAndMasterByAssetDateWithoutHiddenOrDeleted",
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
		),
	)
	if err != nil {
		return nil, err
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

	var photos []Photo
	for _, masterRecord := range masterRecords {
		if asset, exists := assetRecords[masterRecord.RecordName]; exists {
			photos = append(photos, Photo{
				RecordName:   asset.RecordName,
				Fields:       asset.Fields,
				MasterFields: masterRecord.Fields,
			})
		}
	}

	return photos, nil
}

func (p *photoService) getAllPhotos(ctx context.Context) []Photo {
	appctx.AppTrace(ctx)
	defer appctx.DeferAppTrace(ctx)
	progress, ok := appctx.Progress(ctx)

	var (
		allPhotos []Photo
		offset    = int64(0)
	)

	for {
		photos, err := p.getPhotos(ctx, offset)
		if err != nil {
			return nil
		}

		if len(photos) == 0 {
			break
		}

		allPhotos = append(allPhotos, photos...)
		if ok {
			progress.Count("photos_count", float64(len(allPhotos)))
		}

		offset += int64(len(photos))
	}
	return allPhotos
}

func (p *photoService) GetAllPhotos(ctx context.Context) []usecase.Photo {
	appctx.AppTrace(ctx)
	defer appctx.DeferAppTrace(ctx)

	var result []usecase.Photo
	for _, p := range p.getAllPhotos(ctx) {
		v, err := cnvPhoto(p)
		if err != nil {
			slog.ErrorContext(ctx, err.Error()+p.RecordName)
			continue
		}

		result = append(result, v)
	}

	return result
}

func (p *photoService) MakeDownloadUrlByPhotos(ctx context.Context, photos []usecase.Photo) (string, error) {
	appctx.AppTrace(ctx)
	defer appctx.DeferAppTrace(ctx)

	httpClient, _, appleInfo, _ := MetaData(ctx)

	var ids []string
	for _, p := range photos {
		ids = append(ids, p.ID)
	}

	url := util.MustParseUrl(
		appleInfo.WebServiceSckdatabasewsUrl+"/database/1/com.apple.photos.cloud/production/private/records/zip/prepare",
		map[string]string{
			"remapEnums":          "True",
			"getCurrentSyncToken": "True",
		},
	)
	body := bytes.NewBuffer(util.MustMarshal(map[string]any{
		"includeRecords": ids,
		"archiveName":    "icloud" + strconv.FormatInt(time.Now().UnixNano(), 10) + "_" + util.Hash(strings.Join(ids, ",")) + ".zip",
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
		"User-Agent":   util.UserAgent,
	}

	resp, err := util.HttpDoJSON[map[string]any](httpClient, util.MustRequest(ctx, http.MethodPost, url, body, headers))
	if err != nil {
		return "", fmt.Errorf("failed to zip request: %w", err)
	}

	return (*resp)["downloadURL"].(string), nil
}

func cnvPhoto(photo Photo) (usecase.Photo, error) {
	filenameBase64 := photo.MasterFields["filenameEnc"].(map[string]any)["value"].(string)
	decodedBytes, err := base64.StdEncoding.DecodeString(filenameBase64)
	if err != nil {
		return usecase.Photo{}, err
	}

	resOriginalRes := photo.MasterFields["resOriginalRes"].(map[string]any)
	resOriginalResValue := resOriginalRes["value"].(map[string]any)
	fileSize, _ := resOriginalResValue["size"].(float64)
	return usecase.Photo{
		ID:          photo.RecordName,
		CheckSum:    resOriginalResValue["fileChecksum"].(string),
		DownloadUrl: resOriginalResValue["downloadURL"].(string),
		Filename:    string(decodedBytes),
		FileSize:    fileSize,
	}, nil
}
