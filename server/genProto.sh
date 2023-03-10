function genRentalProto {
  DOMAIN=rental
  SKIP_GATEWAY=$1

  # generate go
  PROTO_PATH=./${DOMAIN}/api
  GO_OUT_PATH=./${DOMAIN}/api/gen/v1
  # protoc go & go-grpc
  mkdir -p $GO_OUT_PATH
  protoc -I=$PROTO_PATH rental.proto -I=$PROTO_PATH trip.proto -I=$PROTO_PATH profile.proto --go_out=paths=source_relative:$GO_OUT_PATH ${DOMAIN}.proto;
  protoc -I=$PROTO_PATH rental.proto -I=$PROTO_PATH trip.proto -I=$PROTO_PATH profile.proto --go-grpc_out=paths=source_relative:$GO_OUT_PATH ${DOMAIN}.proto;
  # protoc grpc-gateway
  if [ $SKIP_GATEWAY ]; then
    return
  fi
  protoc -I=$PROTO_PATH rental.proto -I=$PROTO_PATH trip.proto -I=$PROTO_PATH profile.proto --grpc-gateway_out=paths=source_relative,grpc_api_configuration=$PROTO_PATH/rental.yaml:$GO_OUT_PATH ${DOMAIN}.proto;

  # generate javascript
  PBTS_BIN_DIR=../wx/node_modules/.bin
  PBTS_OUT_DIR=../wx/miniprogram/apis/proto_gen/${DOMAIN}
  mkdir -p $PBTS_OUT_DIR
  files=("rental" "trip" "profile")
  for i in ${!files[@]}; do
    # pbjs
    $PBTS_BIN_DIR/pbjs -t static -w es6 $PROTO_PATH/${files[$i]}.proto --no-create --no-encode --no-decode --no-verify --no-delimited --force-number -o $PBTS_OUT_DIR/${files[$i]}_pb_tmp.js
    echo 'import * as $protobuf from "protobufjs";' > $PBTS_OUT_DIR/${files[$i]}_pb.js
    cat $PBTS_OUT_DIR/${files[$i]}_pb_tmp.js >> $PBTS_OUT_DIR/${files[$i]}_pb.js
    rm $PBTS_OUT_DIR/${files[$i]}_pb_tmp.js
    # pbts
    $PBTS_BIN_DIR/pbts -o $PBTS_OUT_DIR/${files[$i]}_pb.d.ts $PBTS_OUT_DIR/${files[$i]}_pb.js
  done
}

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
  PBTS_OUT_DIR=../wx/miniprogram/apis/proto_gen/${DOMAIN}
  # pbjs
  mkdir -p $PBTS_OUT_DIR
  $PBTS_BIN_DIR/pbjs -t static -w es6 $PROTO_PATH/${DOMAIN}.proto --no-create --no-encode --no-decode --no-verify --no-delimited --force-number -o $PBTS_OUT_DIR/${DOMAIN}_pb_tmp.js
  echo 'import * as $protobuf from "protobufjs";' > $PBTS_OUT_DIR/${DOMAIN}_pb.js
  cat $PBTS_OUT_DIR/${DOMAIN}_pb_tmp.js >> $PBTS_OUT_DIR/${DOMAIN}_pb.js
  rm $PBTS_OUT_DIR/${DOMAIN}_pb_tmp.js
  # pbts
  $PBTS_BIN_DIR/pbts -o $PBTS_OUT_DIR/${DOMAIN}_pb.d.ts $PBTS_OUT_DIR/${DOMAIN}_pb.js
}

genProto auth
genRentalProto
