from typing import List

from fastapi import FastAPI

from .boomer_container import Boomer, Slave, CreateSlave, ContainerConfig, get_client


app = FastAPI()


@app.post("/create_slave", response_model=Slave)
def create_slave(c: CreateSlave):
    client = get_client(c.container_config)
    b = Boomer(client=client)
    boomer_cmd = c.boomer_cmd
    return b.create_slave(image=c.image, req_cmd=c.request, boomer_cmd=boomer_cmd)


@app.get("/list_slave", response_model=List[Slave])
def list_slave():
    config = ContainerConfig()
    client = get_client(config)
    b = Boomer(client=client)
    return b.list_slave()


@app.delete("/remove_slave_by_id")
def remove_slave_by_id(container_id: str, force: bool = False):
    config = ContainerConfig()
    client = get_client(config)
    b = Boomer(client=client)
    return b.remove_by_id(container_id, force)


@app.delete("/remove_all_slave")
def remove_all_slave(force: bool = False):
    config = ContainerConfig()
    client = get_client(config)
    b = Boomer(client=client)
    return b.remove_all_slave(force)


@app.post("/stop_slave_by_id")
def stop_slave_by_id(container_id: str):
    config = ContainerConfig()
    client = get_client(config)
    b = Boomer(client=client)
    return b.stop_by_id(container_id)


@app.post("/stop_all_slave")
def stop_all_slave():
    config = ContainerConfig()
    client = get_client(config)
    b = Boomer(client=client)
    return b.stop_all_slave()
