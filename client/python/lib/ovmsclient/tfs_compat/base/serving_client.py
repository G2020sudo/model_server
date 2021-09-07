#
# Copyright (c) 2021 Intel Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

import os
from validators import ipv4, domain
from abc import ABC, abstractmethod


class ServingClient(ABC):

    @abstractmethod
    def predict(self, request):
        '''
        Send PredictRequest to the server and return response.

        Args:
            request: PredictRequest object.

        Returns:
            PredictResponse object

        Raises:
            TypeError:  if provided argument is of wrong type.
            Many more for different serving reponses...
        '''

        pass

    @abstractmethod
    def get_model_metadata(self, request):
        '''
        Send ModelMetadataRequest to the server and return response.

        Args:
            request: ModelMetadataRequest object.

        Returns:
            ModelMetadataResponse object

        Raises:
            TypeError:  if provided argument is of wrong type.
            Many more for different serving reponses...
        '''

        pass

    @abstractmethod
    def get_model_status(self, request):
        '''
        Send ModelStatusRequest to the server and return response.

        Args:
            request: ModelStatusRequest object.

        Returns:
            ModelStatusResponse object

        Raises:
            TypeError:  if provided argument is of wrong type.
            Many more for different serving reponses...
        '''

        pass

    @classmethod
    @abstractmethod
    def _build(cls, config):
        raise NotImplementedError

    @classmethod
    def _prepare_certs(cls, server_cert_path, client_cert_path, client_key_path):

        client_cert, client_key = None, None

        server_cert = cls._open_certificate(server_cert_path)

        if client_cert_path is not None:
            client_cert = cls._open_certificate(client_cert_path)

        if client_key_path is not None:
            client_key = cls._open_private_key(client_key_path)

        return server_cert, client_cert, client_key

    @classmethod
    def _open_certificate(cls, certificate_path):
        with open(certificate_path, 'rb') as f:
            certificate = f.read()
            return certificate

    @classmethod
    def _open_private_key(cls, key_path):
        with open(key_path, 'rb') as f:
            key = f.read()
            return key

    @classmethod
    def _check_config(cls, config):

        if 'address' not in config or 'port' not in config:
            raise ValueError('The minimal config must contain address and port')

        cls._check_address(config['address'])

        cls._check_port(config['port'])

        if 'tls_config' in config:
            cls._check_tls_config(config['tls_config'])

    @classmethod
    def _check_address(cls, address):

        if not isinstance(address, str):
            raise TypeError(f'address type should be string, but is {type(address).__name__}')

        if address != "localhost" and not ipv4(address) and not domain(address):
            raise ValueError('address is not valid')

    @classmethod
    def _check_port(cls, port):

        if not isinstance(port, int):
            raise TypeError(f'port type should be int, but is type {type(port).__name__}')

        if port.bit_length() > 16 or port < 0:
            raise ValueError(f'port should be in range <0, {2**16-1}>')

    @classmethod
    def _check_tls_config(cls, tls_config):

        if 'server_cert_path' not in tls_config:
            raise ValueError('server_cert_path is not defined in tls_config')

        if ('client_key_path' in tls_config) != ('client_cert_path' in tls_config):
            raise ValueError('none or both client_key_path and client_cert_path '
                             'are required in tls_config')

        valid_keys = ['server_cert_path', 'client_key_path', 'client_cert_path']
        for key in tls_config:
            if key not in valid_keys:
                raise ValueError(f'{key} is not valid tls_config key')
            if not isinstance(tls_config[key], str):
                raise TypeError(f'{key} type should be string but is type '
                                f'{type(tls_config[key]).__name__}')
            if not os.path.isfile(tls_config[key]):
                raise ValueError(f'{tls_config[key]} is not valid path to file')
