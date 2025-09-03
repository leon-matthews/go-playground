# Monarch

CLI helper for file conversion tasks

## Tools

### mediainfo

    $ mediainfo --Output=JSON testdata/cow.mp4

    {
      "creatingLibrary": {
        "name": "MediaInfoLib",
        "version": "24.12",
        "url": "https://mediaarea.net/mediaInfo"
      },
      "media": {
        "@ref": "testdata/cow.mp4",
        "track": [
          {
            "@type": "General",
            "VideoCount": "1",
            "AudioCount": "1",
            "FileExtension": "mp4",
            "Format": "MPEG-4",
            "Format_Profile": "Base Media",
            "CodecID": "isom",
            "CodecID_Compatible": "isom/iso2/mp41",
            "FileSize": "483210",
            "Duration": "4.016",
            "OverallBitRate": "962570",
            ...
          },
          {
            "@type": "Video",
            "StreamOrder": "0",
            "ID": "1",
            "Format": "HEVC",
            "Format_Profile": "Main",
            "Format_Level": "3",
            "Format_Tier": "Main",
            "CodecID": "hev1",
            "Duration": "4.000",
            "BitRate": "819440",
            "Width": "480",
            "Height": "848",
            "ColorSpace": "YUV",
            ...
          },
          {
            "@type": "Audio",
            "StreamOrder": "1",
            "ID": "2",
            "Format": "AAC",
            "Source_Duration": "4.011",
            "BitRate_Mode": "CBR",
            "BitRate": "132300",
            "Channels": "2",
            ...
          }
        ]
      }
    }


### ffprobe

An alternative to the `mediainfo` tool:

    $ ffprobe -v quiet -print_format json -show_format -show_streams video.mp4
    {
        "streams": [
            {
                "index": 0,
                "codec_name": "h264",
                ...
                "width": 1280,
                "height": 720,
            },
            {
                "index": 1,
                "codec_name": "aac",
                "sample_rate": "44100",
                "channels": 2,
                "bit_rate": "192000",
                ...
            }
        ],
        "format": {
            "filename": "video.mp4",
            "nb_streams": 2,
            "format_name": "mov,mp4,m4a,3gp,3g2,mj2",
            "format_long_name": "QuickTime / MOV",
            "start_time": "0.000000",
            "duration": "6206.569887",
            "size": "3882420728",
            "bit_rate": "5004272",
            ...
        }
    }
