from typing import List

from fastapi import FastAPI

from .boomer_container import Boomer, Container, CreateContainer, SlaveConfig

app = FastAPI()


@app.post("/create_container", response_model=Container)
def create_container(c: CreateContainer):
    sc = SlaveConfig()
    b = Boomer(slave_config=sc)
    return b.create_container(c.image, c.request)
    # res = b.list_containers()
    # return [{'name': x.name, 'status': x.status, 'tags': [x.image.tags]} for x in res]


@app.get("/list_containers", response_model=List[Container])
def list_containers():
    sc = SlaveConfig()
    b = Boomer(slave_config=sc)
    return b.list_containers()
