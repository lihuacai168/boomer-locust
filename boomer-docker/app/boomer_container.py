import json
from typing import Union, List, Optional, Any

import docker
from docker import DockerClient
from pydantic import BaseModel
from loguru import logger


class Slave(BaseModel):
    id: str
    name: str
    status: str
    image: str
    cmd: List[str] = []
    entrypoint: List[Any] = None


class RequestCmd(BaseModel):
    url: str
    method: str
    headers: dict = {'Content-Type': 'application/json'}
    body: Union[list, dict] = None

    def to_cmd(self):
        base_cmd = f''' --url {self.url} --method {self.method} --json-headers \'{self.headers}\''''
        if self.body is not None:
            base_cmd += f''' --raw-data \'{json.dumps(self.body)}\''''
        return base_cmd


class BoomerCmd(BaseModel):
    master_host: str = '127.0.0.1'
    master_port: int = 5557
    max_rps: int
    request_increase_rate: int = -1
    verbose = False

    def to_cmd(self):
        cmd = f''' --master-host {self.master_host} --master-port {self.master_port} --max-rps {self.max_rps} '''
        return cmd


class ContainerConfig(BaseModel):
    host: Optional[str] = None
    port: Optional[int] = None
    socket: str = 'unix://var/run/docker.sock'


class CreateSlave(BaseModel):
    image: str = 'boomer:latest'
    request: RequestCmd
    container_config: ContainerConfig
    boomer_cmd: BoomerCmd


def get_client(config: ContainerConfig) -> DockerClient:
    if config.host and config.port:
        return docker.DockerClient(base_url=f'tcp://{config.host}:{config.port}')
    return docker.DockerClient(base_url=config.socket)


class Boomer(object):
    def __init__(self, client: DockerClient):
        self.client = client

    def create_slave(self, image, req_cmd: RequestCmd, boomer_cmd: BoomerCmd) -> Slave:

        b_cmd = boomer_cmd.to_cmd()
        # cmd = '''--master-host=192.168.31.12  --url=http://192.168.31.12:8081/post  --method=POST --content-type="application/json"  --raw-data='[123,345]' --verbose 1'''
        r_cmd = req_cmd.to_cmd()
        cmd = b_cmd + r_cmd
        # verbose only works with the end of the command
        if boomer_cmd.verbose:
            cmd += ' --verbose 1'
        container = self.client.containers.run(image, cmd, detach=True)
        data = {
            'id': container.id,
            'name': container.name,
            'status': container.status,
            'image': container.image.tags[0],
        }
        return Slave(**data)

    def stop_by_id(self, container_id):
        return self._do_by_id(container_id, 'stop')

    def remove_by_id(self, container_id: str, force: bool = False) -> bool:
        if force:
            self.stop_by_id(container_id)
        return self._do_by_id(container_id, 'remove')

    def _do_by_id(self, container_id, cmd):
        try:
            container = self.client.containers.get(container_id)
        except Exception as e:
            logger.warning(f'Container {container_id} not found')
            return False

        getattr(container, cmd)()
        return True

    def stop_by_name(self, name):
        return self._do_by_name(name, 'stop')

    def remove_by_name(self, name):
        return self._do_by_name(name, 'remove')

    def _do_by_name(self, name, cmd):
        for container in self.client.containers.list():
            if container.name == name:
                getattr(container, cmd)()
                return True
        return False

    def list_slave(self) -> List[Slave]:
        return [Slave(id=c.id,
                      name=c.name,
                      status=c.status,
                      image=c.image.tags[0],
                      entrypoint=c.attrs['Config']['Entrypoint'],
                      cmd=c.attrs['Config']['Cmd']) for c in self.client.containers.list() if
                c.image.tags[0].startswith('boomer')]

    def stop_all_slave(self):
        for container in self.client.containers.list():
            if container.image.tags[0].startswith('boomer'):
                container.stop()

    def remove_all_slave(self, force: bool = False) -> bool:
        if force:
            self.stop_all_slave()
        for c in self.client.containers.list():
            if c.image.tags[0].startswith('boomer'):
                self.client.containers.get(c.id).stop()
        return True


if __name__ == '__main__':
    boomer = Boomer()
    print(boomer.list_slave())
