#!/bin/bash

# Test GetFindTyre with curl to see raw response
curl -X POST "http://api-b2b.4tochki.ru/WCF/ClientService.svc" \
  -H "Content-Type: text/xml; charset=utf-8" \
  -H "SOAPAction: Wcf.ClientService.Client.WebAPI.TS3/ClientService/GetFindTyre" \
  -d '<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"
                  xmlns:wcf="Wcf.ClientService.Client.WebAPI.TS3"
                  xmlns:ts3="http://schemas.datacontract.org/2004/07/TS3.Domain.Models.Client.ClientSoapService.SearchTires"
                  xmlns:arr="http://schemas.microsoft.com/2003/10/Serialization/Arrays">
  <soapenv:Body>
    <wcf:GetFindTyre>
      <wcf:login>sa69263</wcf:login>
      <wcf:password>jkHP4Nj3)z</wcf:password>
      <wcf:filter></wcf:filter>
      <wcf:page>0</wcf:page>
      <wcf:pageSize>10</wcf:pageSize>
    </wcf:GetFindTyre>
  </soapenv:Body>
</soapenv:Envelope>' | xmllint --format - | head -100
