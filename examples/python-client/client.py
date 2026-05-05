#!/usr/bin/env python3
"""
Python gRPC Client for rpc-gen Go Server
"""

import grpc
import time
from typing import List, Optional
import logging

logging.basicConfig(level=logging.INFO, format='%(asctime)s [%(levelname)s] %(name)s: %(message)s')
logger = logging.getLogger(__name__)

# Generated protobuf imports
try:
    from google.rpc_codegen_api.v1 import rpc_codegen_pb2 as pb2
    from google.rpc_codegen_api.v1 import rpc_codegen_pb2_grpc as grpc_stub
except ImportError:
    logger.warning("Generated stubs not found. Run: protoc --python_out=. api/rpc_codegen.proto")
    class _Msg:
        def __init__(self, **kw):
            for k, v in kw.items(): setattr(self, k, v)
        def __getattr__(self, name):
            if name.startswith("Get_") or name.startswith("get_"):
                return getattr(self, name[4:], None)
            raise AttributeError(name)
    class HelloRequest(_Msg): pass
    class HelloResponse(_Msg): pass
    class HelloStreamRequest(_Msg): pass
    class HelloStreamResponse(_Msg): pass
    class _Stub: pass
    class HelloService_Stub(_Stub): pass
    grpc_stub = type("grpc_stub", (), {"HelloService_Stub": HelloService_Stub})
    pb2 = None

DEFAULT_HOST = "localhost"
DEFAULT_PORT = 50051
REQUEST_TIMEOUT = 10.0


class RetryClient:
    def __init__(self, channel, max_retries=3, backoff=1.0):
        self._channel = channel
        self._max_retries = max_retries
        self._backoff = backoff

    def _retry(self, method, *args, **kwargs):
        for attempt in range(self._max_retries):
            try:
                return method(*args, **kwargs)
            except grpc.RpcError as e:
                if attempt == self._max_retries - 1:
                    raise
                wait = self._backoff * (2 ** attempt)
                logger.warning(f"RPC failed (attempt {attempt + 1}/{self._max_retries}): {e}")
                time.sleep(wait)

    def close(self):
        self._channel.close()


class HelloServiceClient:
    def __init__(self, host=DEFAULT_HOST, port=DEFAULT_PORT, secure=False, max_retries=3):
        if secure:
            channel = grpc.secure_channel(f"{host}:{port}", grpc.ssl_channel_credentials())
        else:
            channel = grpc.insecure_channel(f"{host}:{port}")
        self._channel = channel
        self._stub = grpc_stub.HelloService_Stub(channel)
        self._client = RetryClient(channel, max_retries=max_retries)

    def close(self):
        self._channel.close()

    def hello(self, name, greet_type="greeting"):
        req = pb2.HelloRequest(name=name, greet_type=greet_type) if pb2 else HelloRequest(name=name, greet_type=greet_type)
        try:
            resp = self._client._retry(self._stub.Hello, req, timeout=REQUEST_TIMEOUT)
            logger.info(f"Got: {resp.message} (server: {resp.server_id})")
            return resp
        except grpc.RpcError as e:
            logger.error(f"RPC error: {e.code()}: {e.details()}")
            raise

    def hello_stream(self, name):
        req = pb2.HelloRequest(name=name) if pb2 else HelloRequest(name=name)
        try:
            responses = self._client._retry(self._stub.HelloStream, req, timeout=REQUEST_TIMEOUT)
            result = []
            for r in responses:
                logger.info(f"  Stream #{r.response_index}: {r.message}")
                result.append(r)
            return result
        except grpc.RpcError as e:
            logger.error(f"RPC error: {e.code()}: {e.details()}")
            raise

    def batch_hello(self, names):
        it = (pb2.HelloRequest(name=n) if pb2 else HelloRequest(name=n) for n in names)
        try:
            resp = self._client._retry(self._stub.BatchHello, it, timeout=REQUEST_TIMEOUT)
            logger.info(f"Batch: {resp.message}")
            return resp
        except grpc.RpcError as e:
            logger.error(f"RPC error: {e.code()}: {e.details()}")
            raise

    def hello_stream_stream(self, names):
        it = (pb2.HelloStreamRequest(name=n) if pb2 else HelloStreamRequest(name=n) for n in names)
        try:
            responses = self._client._retry(self._stub.HelloStreamStream, it, timeout=REQUEST_TIMEOUT)
            result = []
            for r in responses:
                logger.info(f"  Echo #{r.response_index}: {r.message}")
                result.append(r)
            return result
        except grpc.RpcError as e:
            logger.error(f"RPC error: {e.code()}: {e.details()}")
            raise


def main():
    logger.info("Connecting to RPC server...")
    client = HelloServiceClient(host="localhost", port=50051)
    logger.info("Connected!")
    try:
        logger.info("--- 1. Unary RPC ---")
        client.hello(name="World")

        logger.info("--- 2. Server-Streaming ---")
        client.hello_stream(name="StreamUser")

        logger.info("--- 3. Client-Streaming ---")
        client.batch_hello(names=["Alice", "Bob", "Charlie"])

        logger.info("--- 4. Bidirectional-Streaming ---")
        client.hello_stream_stream(names=["Echo1", "Echo2", "Echo3"])

        logger.info("All RPC calls completed!")
    except grpc.RpcError as e:
        logger.error(f"RPC error: {e.code()}: {e.details()}")
    finally:
        client.close()


if __name__ == "__main__":
    main()
