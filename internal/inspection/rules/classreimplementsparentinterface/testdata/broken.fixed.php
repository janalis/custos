<?php
interface Shape {}
class Base implements Shape {}
// Recovered list with an empty entry: nothing is reported for it.
class Square extends Base implements  {}
