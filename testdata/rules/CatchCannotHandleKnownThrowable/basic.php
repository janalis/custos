<?php
try { throw new Error('failed'); } catch (<warning descr="Catch a type that handles the thrown value.">Exception</warning> $e) {}
try { throw new Error; } catch (<warning descr="Catch a type that handles the thrown value.">RuntimeException | LogicException</warning> $e) {}
try { throw new TypeError($message, 3, null); } catch (<warning descr="Catch a type that handles the thrown value.">Exception</warning> $e) {}
try { throw new Error('failed'); } catch (<warning descr="Catch a type that handles the thrown value.">TypeError</warning> $e) {}
