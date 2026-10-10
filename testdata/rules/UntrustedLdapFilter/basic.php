<?php
ldap_search($ldap,$base,<warning descr="Escape request data for LDAP filter syntax.">"(uid=".$_GET["user"].")"</warning>);
