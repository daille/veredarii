#  ![Veredarii red de interoperabilidad]

##  Instalación de Veredarii

1.- Descargue la última versión de Veredarii desde [GitHub](https://github.com/jcdaille/Veredarii/releases).

2.- Abra una terminal

3.- Navegue hasta el directorio donde descargó el binario

4.- Unase a la red

    Necesita una invitación y la llave de la red, que debe haber recibido de una entidad ya integrada a la red
```bash
./veredarii join --network <name> --entity <entityName> --invitation <invitation.vni> --inviter <entityName> --key <key>
```

5.- Ejecute el binario Veredarii

```bash
./veredarii start
```

6.- Ya está conectado a la red y puede acceder a los recursos que le hayan compartido otras entidades
