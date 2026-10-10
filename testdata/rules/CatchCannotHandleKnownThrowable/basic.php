<?php
try { throw new Error('failed'); } catch (<warning descr="Catch a type that handles the thrown value.">Exception</warning> $e) {}
