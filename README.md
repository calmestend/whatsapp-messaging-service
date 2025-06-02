# Whatsapp Messaging Service

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
    "from": "463752660158443",
    "to": "524437287686"
    }'
    ```


* POST /api/v1/template/encuesta_nps
    ```bash 
    curl -X POST http://localhost:8080/api/v1/template/encuesta_nps \
    -H "Content-Type: application/json" \
    -d '{
    "address": "Jaspers Market, 123 Baker Street. Palo Alto CA, 91041",
    "date": "1 de enero de 2024",
    "from": "463752660158443",
    "to": "524437287686"
    }'
    ```

* POST /api/v1/template/pedido_confirmado
    ```bash 
    curl -X POST http://localhost:8080/api/v1/template/pedido_confirmado \
    -F "to=524437287686" \
    -F "from=463752660158443" \
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
    -F "to=524437287686" \
    -F "from=463752660158443" \
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
    -F "to=524437287686" \
    -F "from=463752660158443" \
    -F "supplierName=Muebles X" \
    -F "businessName=Muebleria Y" \
    -F "file=@hello.pdf"
    ```
