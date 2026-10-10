<?php
class Meter {public int $value;} echo <error descr="Initialize this typed property before reading it.">(new Meter())->value</error>;
