function genProto {
  DOMAIN=$1
  SKIP_GATEWAY=$2

  # generate go
  PROTO_PATH=./${DOMAIN}/api
  GO_OUT_PATH=./${DOMAIN}/api/gen/v1
  # protoc go & go-grpc
  mkdir -p $GO_OUT_PATH
  protoc -I=$PROTO_PATH --go_out=paths=source_relative:$GO_OUT_PATH ${DOMAIN}.proto;
  protoc -I=$PROTO_PATH --go-grpc_out=paths=source_relative:$GO_OUT_PATH ${DOMAIN}.proto;
  # protoc grpc-gateway
  if [ $SKIP_GATEWAY ]; then
    return
  fi
  protoc -I=$PROTO_PATH --grpc-gateway_out=paths=source_relative,grpc_api_configuration=$PROTO_PATH/${DOMAIN}.yaml:$GO_OUT_PATH ${DOMAIN}.proto;

  # generate javascript
  PBTS_BIN_DIR=../wx/node_modules/.bin
  PBTS_OUT_DIR=../wx/miniprogram/services/proto_gen/${DOMAIN}
  # pbjs
  mkdir -p $PBTS_OUT_DIR
  $PBTS_BIN_DIR/pbjs -t static -w es6 $PROTO_PATH/${DOMAIN}.proto --no-create --no-encode --no-decode --no-verify --no-delimited -o $PBTS_OUT_DIR/${DOMAIN}_pb_tmp.js
  echo 'import * as $protobuf from "protobufjs";' > $PBTS_OUT_DIR/${DOMAIN}_pb.js
  cat $PBTS_OUT_DIR/${DOMAIN}_pb_tmp.js >> $PBTS_OUT_DIR/${DOMAIN}_pb.js
  rm $PBTS_OUT_DIR/${DOMAIN}_pb_tmp.js
  # pbts
  $PBTS_BIN_DIR/pbts -o $PBTS_OUT_DIR/${DOMAIN}_pb.d.ts $PBTS_OUT_DIR/${DOMAIN}_pb.js
}

genProto auth
genProto rental
