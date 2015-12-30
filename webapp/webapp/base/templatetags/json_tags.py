# -*- coding: utf-8 -*-
import json

from django import template
from django.utils.html import escapejs, escape
from django.utils.safestring import mark_safe

register = template.Library()


@register.filter
def to_js_zone(collection):
    mapped_items = []
    for item in collection:
        mapped_items.append({'id': item.pk, 'name': item.name})
    return mark_safe('JSON.parse("%s")' % escapejs(json.dumps(mapped_items, separators=(',', ':'))))
