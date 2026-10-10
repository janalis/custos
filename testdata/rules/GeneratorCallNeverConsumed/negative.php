<?php
function deliver() { yield sendPacket(); } foreach (deliver() as $v) {}
