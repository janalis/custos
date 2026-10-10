<?php
array_walk($names,function(&$name){$name=strtoupper($name);}); array_walk($names,function($name){echo $name;});
