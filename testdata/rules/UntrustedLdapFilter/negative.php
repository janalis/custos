<?php
ldap_search($ldap,$base,"(uid=".ldap_escape($_GET["user"],"",LDAP_ESCAPE_FILTER).")");
