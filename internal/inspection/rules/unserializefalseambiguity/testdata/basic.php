<?php
if(<warning descr="Distinguish valid serialized false from decoding failure.">unserialize($bytes,["allowed_classes"=>false])</warning>===false){throw new RuntimeException();}
