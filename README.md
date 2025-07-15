# Whatsapp Messaging Service

## Templates Names
* pagos_atrasados
* encuesta_nps
* pedido_confirmado
* envio_cotizacion
* envio_compra
* cierre_diario

## Endpoints

* POST /api/v1/template/pagos_atrasados
    ```bash 
    curl -X POST http://localhost:8080/api/v1/template/pagos_atrasados \
    -H "Content-Type: application/json" \
    -d '{
    "item": "folio",
    "amount": "12.34",
    "days": "3",
    "condition": "Evitar cargos por retraso",
    "from": "xxx",
    "to": "xxx"
    }'
    ```


* POST /api/v1/template/encuesta_nps
    ```bash 
    curl -X POST http://localhost:8080/api/v1/template/encuesta_nps \
    -H "Content-Type: application/json" \
    -d '{
    "businessName": "Muebleria Juanito",
    "date": "1 de enero de 2024",
    "url": "Encuestas/xxx/xxx.aspx?Id=xxxx"
    "from": "xxx",
    "to": "xxx"
    }'
    ```

* POST /api/v1/template/pedido_confirmado
    ```bash 
    curl -X POST http://localhost:8080/api/v1/template/pedido_confirmado \
    -F "to=xxx" \
    -F "from=xxx" \
    -F "customerName=Alex Velasco" \
    -F "businessName=Sistemas Cubicos" \
    -F "folio=2M2287" \
    -F "amount=5000" \
    -F "numberOfProducts=4" \
    -F "sellerName=Juan Perez" \
    -F "file=@hello.pdf"
    ```


* POST /api/v1/template/envio_cotizacion
    ```bash 
    curl -X POST http://localhost:8080/api/v1/template/envio_cotizacion \
    -F "to=xxx" \
    -F "from=xxx" \
    -F "customerName=Alex Velasco" \
    -F "businessName=Sistemas Cubicos" \
    -F "folio=2M2287" \
    -F "amount=5000" \
    -F "date=28/05/2025" \
    -F "file=@hello.pdf"
    ```

* POST /api/v1/template/envio_compra
    ```bash 
    curl -X POST http://localhost:8080/api/v1/template/envio_compra \
    -F "to=xxx" \
    -F "from=xxx" \
    -F "supplierName=Muebles X" \
    -F "businessName=Muebleria Y" \
    -F "file=@hello.pdf"
    ```

* POST /api/v1/template/cierre_diario
    ```bash 
    curl -X POST http://localhost:8080/api/v1/template/cierre_diario \
      -F "to=xxx" \
      -F "from=xxx" \
      -F "customerName=Alejandro Velazco" \
      -F "businessName=muebleria x" \
      -F "date=09/06/2005" \
      -F "hour=7:23am" \
      -F "file=@hello.pdf"
    ```



* POST /api/v1/template/create/<template_name>
    ```bash 
    curl -X POST http://localhost:8080/api/v1/template/create/<template_name> \
        -F "phone_id=xxx" \
        -F "wba_id=xxx" \
        -F "app_id=xxx" \
        -F "token=xxx" \
        -F "file=@file.pdf"
    ```

* POST /api/v1/templates/create
    ```bash 
    curl -X POST http://localhost:8080/api/v1/templates/create \
        -F "phone_id=xxx" \
        -F "wba_id=xxx" \
        -F "app_id=xxx" \
        -F "token=xxx" \
        -F "file=@file.pdf"
    ```
