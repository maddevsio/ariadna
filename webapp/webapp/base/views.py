# -*- coding: utf-8 -*-
""" Views for the base application """

from django.shortcuts import render
from django.http import JsonResponse
from .models import Address, City, District, Intersection

def home(request):
    """ Default view for the root """
    # addresses = District.objects.all()
    # addresses = Address.objects.filter(city=u'Бишкек').exclude(street="", housenumber="")[65555:]
    addresses = Intersection.objects.exclude(name=' ')[18000:19000]
    return render(request, 'base/home.html', locals())
