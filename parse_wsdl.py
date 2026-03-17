#!/usr/bin/env python3
import xml.etree.ElementTree as ET

# Parse WSDL
tree = ET.parse('4tochki.wsdl')
root = tree.getroot()

# Namespaces
ns = {
    'wsdl': 'http://schemas.xmlsoap.org/wsdl/',
    'soap': 'http://schemas.xmlsoap.org/wsdl/soap/',
    'xsd': 'http://www.w3.org/2001/XMLSchema'
}

print("=" * 60)
print("4tochki WSDL Analysis")
print("=" * 60)
print()

# Find target namespace
target_ns = root.attrib.get('targetNamespace', 'N/A')
print(f"Target Namespace: {target_ns}")
print()

# Find operations in portType
print("Available Operations:")
print("-" * 60)

for portType in root.findall('.//wsdl:portType', ns):
    port_name = portType.attrib.get('name', 'Unknown')
    print(f"\nPort Type: {port_name}")

    for operation in portType.findall('wsdl:operation', ns):
        op_name = operation.attrib.get('name', 'Unknown')
        print(f"  - {op_name}")

        # Find input/output messages
        input_msg = operation.find('wsdl:input', ns)
        output_msg = operation.find('wsdl:output', ns)

        if input_msg is not None:
            input_name = input_msg.attrib.get('message', 'N/A')
            print(f"    Input:  {input_name}")

        if output_msg is not None:
            output_name = output_msg.attrib.get('message', 'N/A')
            print(f"    Output: {output_name}")

print()
print("=" * 60)

# Find SOAP actions
print("\nSOAP Actions:")
print("-" * 60)

for binding in root.findall('.//wsdl:binding', ns):
    binding_name = binding.attrib.get('name', 'Unknown')

    for operation in binding.findall('wsdl:operation', ns):
        op_name = operation.attrib.get('name', 'Unknown')
        soap_operation = operation.find('soap:operation', ns)

        if soap_operation is not None:
            action = soap_operation.attrib.get('soapAction', 'N/A')
            print(f"{op_name}: {action}")

print()
print("=" * 60)
