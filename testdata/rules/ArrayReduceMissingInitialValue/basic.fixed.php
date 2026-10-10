<?php
array_reduce($parts, fn($a,$v)=>array_merge($a,[$v]), []);
