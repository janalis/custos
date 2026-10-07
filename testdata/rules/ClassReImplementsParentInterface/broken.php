<?php
interface Shape {}
class Base implements Shape {}
// Recovered list with an empty entry: nothing is reported for it.
class Square extends Base implements , <warning descr="'\Shape' is already implemented by '\Base'; remove it here.">Shape</warning> {}
