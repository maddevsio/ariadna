""" Basic models, such as user profile """
from django.contrib.gis.db import models

class Address(models.Model):
    city = models.CharField(max_length=255)
    housenumber = models.CharField(max_length=255)
    street = models.CharField(max_length=255)
    coords = models.PolygonField()

    class Meta:
        db_table = 'addresses'


class City(models.Model):
    name = models.CharField(max_length=255)
    country = models.CharField(max_length=255)
    coords = models.PolygonField()

    class Meta:
        db_table = 'cities'


class District(models.Model):
    name = models.CharField(max_length=255)
    coords = models.PolygonField()

    class Meta:
        db_table = 'district'


class Intersection(models.Model):
    name = models.CharField(max_length=255)
    coords = models.GeometryField()

    class Meta:
        db_table = 'road_intersection'


