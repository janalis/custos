<?php
function deliver() { yield sendPacket(); } <warning descr="Consume the generator to execute its body.">deliver()</warning>;
