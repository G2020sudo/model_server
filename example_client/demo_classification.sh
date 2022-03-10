#!/bin/bash


docker run --entrypoint "/bin/bash" --net host ovms-classification:1.0 -c "cd /example_client && python3 grpc_binary_client.py --grpc_port 9000 --input_name sub/placeholder_port_0 --output_name efficientnet-b0/model/head/dense/BiasAdd/Add --model_name efficientnet-b0 --batchsize 1"





