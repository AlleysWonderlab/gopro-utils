echo "--- GoPro to Gargarmel ---"

if [ "$#" -ne 2 ]; then
    echo "Illegal number of parameters"
    exit 0;
fi


TRACK_BIN="tem_track.bin"
GPX_OUT="temp_out.gpx"

echo "-Save meta track in $1 file to $TRACK_BIN"
ffmpeg -y -i $1 -codec copy -map 0:3 -f rawvideo $TRACK_BIN

echo "-Extract gpx information in $TRACK_BIN"
./gopro2gpx -v $1 -i $TRACK_BIN -o $GPX_OUT

echo "-Run gargamel.js"
node gargamel.js -c ../../alleys-server/config.local/ $2 $1 $GPX_OUT

echo "--- End ---"
