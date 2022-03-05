from typing import List

from fastapi import FastAPI, APIRouter

from .boomer_container import Boomer, Slave, CreateSlave, ContainerConfig, get_client

router = APIRouter(prefix="/api/v1")


@router.post("/create_slave", response_model=Slave)
def create_slave(c: CreateSlave):
    client = get_client(c.container_config)
    b = Boomer(client=client)
    boomer_cmd = c.boomer_cmd
    return b.create_slave(image=c.image, req_cmd=c.request, boomer_cmd=boomer_cmd)


@router.get("/list_slave", response_model=List[Slave])
def list_slave():
    config = ContainerConfig()
    client = get_client(config)
    b = Boomer(client=client)
    return b.list_slave()


@router.delete("/remove_slave_by_id")
def remove_slave_by_id(container_id: str, force: bool = False):
    config = ContainerConfig()
    client = get_client(config)
    b = Boomer(client=client)
    return b.remove_by_id(container_id, force)


@router.delete("/remove_all_slave")
def remove_all_slave(force: bool = False):
    config = ContainerConfig()
    client = get_client(config)
    b = Boomer(client=client)
    return b.remove_all_slave(force)


@router.post("/stop_slave_by_id")
def stop_slave_by_id(container_id: str):
    config = ContainerConfig()
    client = get_client(config)
    b = Boomer(client=client)
    return b.stop_by_id(container_id)


@router.post("/stop_all_slave")
def stop_all_slave():
    config = ContainerConfig()
    client = get_client(config)
    b = Boomer(client=client)
    return b.stop_all_slave()


app = FastAPI()
app.include_router(router)
