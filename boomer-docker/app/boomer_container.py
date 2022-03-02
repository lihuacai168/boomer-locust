from typing import Union, List

import docker
from pydantic import BaseModel


class Container(BaseModel):
    id: str
    name: str
    status: str
    image: str


class RequestCmd(BaseModel):
    url: str
    method: str
    headers: dict = {'Content-Type': 'application/json'}
    body: Union[list, dict]

    def to_request_str(self):
        base_cmd = f"--url {self.url} --method {self.method} --headers {self.headers}"
        if self.body:
            base_cmd += f" --raw-data {self.body}"
        return base_cmd


class CreateContainer(BaseModel):
    image: str
    request: RequestCmd


class SlaveConfig(BaseModel):
    host: str = '127.0.0.1'
    port: int = 2375


class Boomer(object):
    def __init__(self, slave_config: SlaveConfig):
        # self.client = docker.from_env()
        # self.client = docker.DockerClient(base_url='unix://var/run/docker.sock')
        self.client = docker.DockerClient(base_url=f'tcp://{slave_config.host}:{slave_config.port}')
        # self.container = self.client.containers.run('boomer:latest', cmd, detach=True)
        # self.container = self.client.containers.run('boomer:latest', detach=True)

    def create_container(self, image, source_req: RequestCmd) -> Container:

        # cmd = '''--master-host=192.168.31.12  --url=http://192.168.31.12:8081/post  --method=POST --content-type="application/json"  --raw-data='[123,345]' --verbose 1'''
        cmd = source_req.to_request_str()

        container = self.client.containers.run(image, cmd, detach=True)
        data = {
            'id': container.id,
            'name': container.name,
            'status': container.status,
            'image': container.image.tags[0],
        }
        return Container(**data)

    def stop_by_id(self, container_id):
        return self._do_by_id(container_id, 'stop')

    def remove_by_id(self, container_id):
        return self._do_by_id(container_id, 'remove')

    def _do_by_id(self, container_id, cmd):
        container = self.client.containers.get(container_id)
        if not container:
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

    def list_containers(self) -> List[Container]:
        # return self.client.containers.list()
        return [Container(id=c.id, name=c.name, status=c.status, image=c.image.tags[0]) for c in self.client.containers.list()]


if __name__ == '__main__':
    boomer = Boomer()
    print(boomer.list_containers())
