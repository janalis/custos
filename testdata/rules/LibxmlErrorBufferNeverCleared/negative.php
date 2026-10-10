<?php
libxml_use_internal_errors(true); $d=new DOMDocument(); foreach($docs as $xml){$d->loadXML($xml);libxml_clear_errors();}
