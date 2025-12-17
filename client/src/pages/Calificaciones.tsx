import React, { useState, useEffect } from 'react';
import {
  Container,
  Paper,
  Typography,
  Grid,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  TextField,
  Button,
  Table,
  TableHead,
  TableBody,
  TableRow,
  TableCell,
  Box,
  Chip,
  LinearProgress,
  CircularProgress,
  Card,
  CardContent,
  Alert,
  IconButton,
  Tooltip
} from '@mui/material';
import {
  Search as SearchIcon,
  FilterList as FilterIcon,
  BarChart as BarChartIcon,
  Refresh as RefreshIcon
} from '@mui/icons-material';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip as ChartTooltip,
  Legend,
  ResponsiveContainer,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell
} from 'recharts';

import {
  obtenerCarreras,
  calcularPromediosAlumnos,
  obtenerEstadisticasCalificaciones,
  type PromedioAlumno
} from '../services/evaluacionesCompletas';

const Calificaciones = () => {
  const [carreras, setCarreras] = useState<any[]>([]);
  const [carreraSeleccionada, setCarreraSeleccionada] = useState<number | ''>('');
  const [alumnos, setAlumnos] = useState<PromedioAlumno[]>([]);
  const [estadisticas, setEstadisticas] = useState<any>(null);
  const [filtroNombre, setFiltroNombre] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    cargarCarreras();
  }, []);

  useEffect(() => {
    if (carreraSeleccionada) {
      cargarDatos();
    } else {
      setAlumnos([]);
      setEstadisticas(null);
    }
  }, [carreraSeleccionada]);

  const cargarCarreras = async () => {
    try {
      const carrerasData = await obtenerCarreras();
      setCarreras(carrerasData);
    } catch (error) {
      console.error('Error cargando carreras:', error);
      setError('Error al cargar las carreras');
    }
  };

  const cargarDatos = async () => {
    if (!carreraSeleccionada) return;

    setLoading(true);
    try {
      const promedios = await calcularPromediosAlumnos(carreraSeleccionada as number);
      setAlumnos(promedios);

      const stats = await obtenerEstadisticasCalificaciones(carreraSeleccionada as number);
      setEstadisticas(stats);
    } catch (error) {
      console.error('Error cargando datos:', error);
      setError('Error al cargar los datos de calificaciones');
    } finally {
      setLoading(false);
    }
  };

  const getColorByGrade = (grade: number) => {
    if (grade >= 9) return 'success';
    if (grade >= 7) return 'primary';
    if (grade >= 6) return 'warning';
    return 'error';
  };

  const getEstadoTexto = (grade: number) => {
    if (grade >= 9) return 'Excelente';
    if (grade >= 7) return 'Bueno';
    if (grade >= 6) return 'Suficiente';
    return 'Reprobado';
  };

  const alumnosFiltrados = alumnos.filter(alumno =>
    alumno.student.Account?.Name?.toLowerCase().includes(filtroNombre.toLowerCase()) ||
    alumno.student.StudentID?.toLowerCase().includes(filtroNombre.toLowerCase())
  );

  const chartData = alumnosFiltrados.map(alumno => ({
    name: alumno.student.Account?.Name?.split(' ')[0] || 'Alumno',
    promedio: alumno.average,
    materias: alumno.subjects_count
  }));

  const distribucionData = estadisticas?.distribucion ?
    Object.entries(estadisticas.distribucion).map(([rango, count]) => ({
      name: rango,
      value: count
    })) : [];

  const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884D8'];

  return (
    <Container maxWidth="xl">
      <Typography variant="h4" gutterBottom sx={{ mb: 3 }}>
        📊 Calificaciones y Promedios
      </Typography>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError('')}>
          {error}
        </Alert>
      )}

      <Paper sx={{ p: 3, mb: 3 }}>
        <Grid container spacing={3} alignItems="center">
          <Grid item xs={12} md={4}>
            <FormControl fullWidth>
              <InputLabel>Carrera</InputLabel>
              <Select
                value={carreraSeleccionada}
                onChange={(e) => setCarreraSeleccionada(e.target.value as number)}
                label="Carrera"
              >
                <MenuItem value="">Todas las carreras</MenuItem>
                {carreras.map((carrera) => (
                  <MenuItem key={carrera.ID} value={carrera.ID}>
                    {carrera.Name}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
          </Grid>

          <Grid item xs={12} md={4}>
            <TextField
              fullWidth
              label="Buscar alumno"
              value={filtroNombre}
              onChange={(e) => setFiltroNombre(e.target.value)}
              placeholder="Nombre o matrícula"
              InputProps={{
                startAdornment: <SearchIcon sx={{ mr: 1, color: 'text.secondary' }} />
              }}
            />
          </Grid>

          <Grid item xs={12} md={4}>
            <Box sx={{ display: 'flex', gap: 1 }}>
              <Button
                variant="outlined"
                startIcon={<FilterIcon />}
                onClick={cargarDatos}
                disabled={!carreraSeleccionada || loading}
                fullWidth
              >
                Aplicar Filtros
              </Button>
              <Tooltip title="Actualizar datos">
                <IconButton
                  onClick={cargarDatos}
                  disabled={!carreraSeleccionada || loading}
                  color="primary"
                >
                  <RefreshIcon />
                </IconButton>
              </Tooltip>
            </Box>
          </Grid>
        </Grid>
      </Paper>

      {estadisticas && (
        <Grid container spacing={3} sx={{ mb: 3 }}>
          <Grid item xs={12} md={3}>
            <Card>
              <CardContent>
                <Typography color="text.secondary" gutterBottom>
                  Promedio General
                </Typography>
                <Typography variant="h4" color="primary.main">
                  {estadisticas.promedioGeneral.toFixed(2)}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  sobre 10 puntos
                </Typography>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} md={3}>
            <Card>
              <CardContent>
                <Typography color="text.secondary" gutterBottom>
                  Total de Alumnos
                </Typography>
                <Typography variant="h4">
                  {estadisticas.totalAlumnos}
                </Typography>
                <Box sx={{ display: 'flex', gap: 1, mt: 1 }}>
                  <Chip
                    label={`${estadisticas.aprobados} aprobados`}
                    color="success"
                    size="small"
                  />
                  <Chip
                    label={`${estadisticas.reprobados} reprobados`}
                    color="error"
                    size="small"
                  />
                </Box>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} md={3}>
            <Card>
              <CardContent>
                <Typography color="text.secondary" gutterBottom>
                  Porcentaje de Aprobación
                </Typography>
                <Typography variant="h4" color="success.main">
                  {estadisticas.totalAlumnos > 0
                    ? `${((estadisticas.aprobados / estadisticas.totalAlumnos) * 100).toFixed(1)}%`
                    : '0%'
                  }
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  de alumnos aprobados
                </Typography>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} md={3}>
            <Card>
              <CardContent>
                <Typography color="text.secondary" gutterBottom>
                  Materias por Alumno
                </Typography>
                <Typography variant="h4">
                  {alumnos.length > 0
                    ? (alumnos.reduce((sum, a) => sum + a.subjects_count, 0) / alumnos.length).toFixed(1)
                    : '0'
                  }
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  promedio de materias
                </Typography>
              </CardContent>
            </Card>
          </Grid>
        </Grid>
      )}

      {loading ? (
        <Box sx={{ display: 'flex', justifyContent: 'center', my: 4 }}>
          <CircularProgress />
        </Box>
      ) : (
        <>
          {alumnos.length > 0 && (
            <Grid container spacing={3} sx={{ mb: 3 }}>
              <Grid item xs={12} md={6}>
                <Paper sx={{ p: 2 }}>
                  <Typography variant="h6" gutterBottom sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                    <BarChartIcon /> Promedios por Alumno
                  </Typography>
                  <ResponsiveContainer width="100%" height={300}>
                    <BarChart data={chartData}>
                      <CartesianGrid strokeDasharray="3 3" />
                      <XAxis dataKey="name" />
                      <YAxis domain={[0, 10]} />
                      <ChartTooltip />
                      <Legend />
                      <Bar
                        dataKey="promedio"
                        fill="#8884d8"
                        name="Promedio"
                        radius={[4, 4, 0, 0]}
                      />
                    </BarChart>
                  </ResponsiveContainer>
                </Paper>
              </Grid>
              <Grid item xs={12} md={6}>
                <Paper sx={{ p: 2 }}>
                  <Typography variant="h6" gutterBottom>
                    Distribución de Calificaciones
                  </Typography>
                  <ResponsiveContainer width="100%" height={300}>
                    <PieChart>
                      <Pie
                        data={distribucionData}
                        cx="50%"
                        cy="50%"
                        labelLine={false}
                        label={({ name, percent }) => `${name}: ${(percent * 100).toFixed(0)}%`}
                        outerRadius={80}
                        fill="#8884d8"
                        dataKey="value"
                      >
                        {distribucionData.map((entry, index) => (
                          <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                        ))}
                      </Pie>
                      <ChartTooltip />
                    </PieChart>
                  </ResponsiveContainer>
                </Paper>
              </Grid>
            </Grid>
          )}

          {alumnosFiltrados.length > 0 && (
            <Paper>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>Matrícula</TableCell>
                    <TableCell>Nombre</TableCell>
                    <TableCell>Semestre</TableCell>
                    <TableCell>Promedio</TableCell>
                    <TableCell>Estado</TableCell>
                    <TableCell>Materias</TableCell>
                    <TableCell>Detalle de Calificaciones</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {alumnosFiltrados.map((alumno) => (
                    <TableRow key={alumno.student.ID} hover>
                      <TableCell>
                        <Typography variant="body2" fontWeight="medium">
                          {alumno.student.StudentID}
                        </Typography>
                      </TableCell>
                      <TableCell>{alumno.student.Account?.Name}</TableCell>
                      <TableCell>
                        <Chip
                          label={`${alumno.student.Semester || 1}°`}
                          size="small"
                          color="primary"
                          variant="outlined"
                        />
                      </TableCell>
                      <TableCell>
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                          <Chip
                            label={alumno.average.toFixed(2)}
                            color={getColorByGrade(alumno.average)}
                            variant="outlined"
                          />
                          <Typography variant="body2" color="text.secondary">
                            {getEstadoTexto(alumno.average)}
                          </Typography>
                        </Box>
                      </TableCell>
                      <TableCell>
                        <Chip
                          label={alumno.average >= 6 ? 'Aprobado' : 'Reprobado'}
                          color={alumno.average >= 6 ? 'success' : 'error'}
                          size="small"
                        />
                      </TableCell>
                      <TableCell>
                        <Typography>
                          {alumno.subjects_count} materias calificadas
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5, maxWidth: 300 }}>
                          {alumno.details.map((detalle, idx) => (
                            <Chip
                              key={idx}
                              label={`${detalle.subject_name}: ${detalle.grade.toFixed(1)}`}
                              size="small"
                              variant="outlined"
                              color={getColorByGrade(detalle.grade)}
                            />
                          ))}
                          {alumno.details.length === 0 && (
                            <Typography variant="body2" color="text.secondary">
                              Sin calificaciones registradas
                            </Typography>
                          )}
                        </Box>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </Paper>
          )}
        </>
      )}

      {!loading && alumnosFiltrados.length === 0 && carreraSeleccionada && (
        <Paper sx={{ p: 3, textAlign: 'center' }}>
          <Typography color="text.secondary">
            No se encontraron alumnos con calificaciones en los filtros seleccionados
          </Typography>
        </Paper>
      )}
    </Container>
  );
};

export default Calificaciones;
