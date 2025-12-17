import React, { useState, useEffect } from 'react';
import {
  Container, Grid, Paper, Typography, FormControl, InputLabel, Select,
  MenuItem, TextField, Button, Table, TableHead, TableBody, TableRow,
  TableCell, Box, Chip, Tooltip, CircularProgress, Snackbar, Alert
} from '@mui/material';
import {
  Save as SaveIcon,
  SaveAlt as SaveAllIcon,
  FilterList as FilterIcon
} from '@mui/icons-material';
import {
  obtenerCarreras,
  obtenerGruposDisponibles,
  obtenerMateriasPorCarrera,
  obtenerAlumnosParaCalificar,
  guardarCalificacionAlumno,
  guardarCalificacionesMasivo,
  type AlumnoParaCalificar
} from '../services/evaluacionesCompletas';

interface Grupo {
  id: string;
  nombre: string;
}

const Evaluaciones = () => {
  const [carreras, setCarreras] = useState<any[]>([]);
  const [carreraSeleccionada, setCarreraSeleccionada] = useState<number | ''>('');
  const [grupos, setGrupos] = useState<Grupo[]>([]);
  const [grupoSeleccionado, setGrupoSeleccionado] = useState<string | ''>('');
  const [materias, setMaterias] = useState<any[]>([]);
  const [materiaSeleccionada, setMateriaSeleccionada] = useState<number | ''>('');
  const [alumnos, setAlumnos] = useState<AlumnoParaCalificar[]>([]);
  const [calificaciones, setCalificaciones] = useState<Record<number, string>>({});
  const [loading, setLoading] = useState(false);
  const [loadingFiltros, setLoadingFiltros] = useState(false);
  const [snackbar, setSnackbar] = useState({ open: false, message: '', severity: 'success' as 'success' | 'error' });

  useEffect(() => {
    cargarCarreras();
  }, []);

  const cargarCarreras = async () => {
    try {
      setLoadingFiltros(true);
      const carrerasData = await obtenerCarreras();
      setCarreras(carrerasData);

      const gruposData = obtenerGruposDisponibles().map((g) => ({ id: g, nombre: g }));
      setGrupos(gruposData);
    } catch (error) {
      console.error('Error cargando carreras:', error);
      mostrarError('Error al cargar las carreras');
    } finally {
      setLoadingFiltros(false);
    }
  };

  useEffect(() => {
    if (carreraSeleccionada) {
      cargarMaterias(carreraSeleccionada);
    } else {
      setMaterias([]);
      setMateriaSeleccionada('');
    }
  }, [carreraSeleccionada]);

  const cargarMaterias = async (careerId: number) => {
    try {
      setLoadingFiltros(true);
      const materiasData = await obtenerMateriasPorCarrera(careerId);
      setMaterias(materiasData);
    } catch (error) {
      console.error('Error cargando materias:', error);
      mostrarError('Error al cargar las materias');
    } finally {
      setLoadingFiltros(false);
    }
  };

  useEffect(() => {
    if (carreraSeleccionada && grupoSeleccionado && materiaSeleccionada) {
      cargarAlumnos();
    } else {
      setAlumnos([]);
      setCalificaciones({});
    }
  }, [carreraSeleccionada, grupoSeleccionado, materiaSeleccionada]);

  const cargarAlumnos = async () => {
    try {
      setLoading(true);
      const alumnosData = await obtenerAlumnosParaCalificar(
        carreraSeleccionada as number,
        materiaSeleccionada as number
      );
      setAlumnos(alumnosData);

      const initialGrades: Record<number, string> = {};
      alumnosData.forEach(a => {
        initialGrades[a.id] = a.currentGrade?.toString() || '';
      });
      setCalificaciones(initialGrades);
    } catch (error) {
      console.error('Error cargando alumnos:', error);
      mostrarError('Error al cargar los alumnos');
    } finally {
      setLoading(false);
    }
  };

  const handleCalificacionChange = (studentId: number, value: string) => {
    const numValue = parseFloat(value);
    if (value && (isNaN(numValue) || numValue < 0 || numValue > 10)) {
      mostrarError('La calificación debe ser entre 0 y 10');
      return;
    }
    setCalificaciones(prev => ({ ...prev, [studentId]: value }));
  };

  const guardarCalificacionIndividual = async (alumno: AlumnoParaCalificar) => {
    const calificacion = calificaciones[alumno.id];
    if (!calificacion || calificacion.trim() === '') {
      mostrarError('Ingresa una calificación primero');
      return;
    }
    const gradeValue = parseFloat(calificacion);
    if (isNaN(gradeValue) || gradeValue < 0 || gradeValue > 10) {
      mostrarError('La calificación debe ser entre 0 y 10');
      return;
    }

    try {
      const success = await guardarCalificacionAlumno(
        alumno.id,
        materiaSeleccionada as number,
        gradeValue,
        `Calificación asignada - ${new Date().toLocaleDateString()}`
      );
      if (success) {
        mostrarMensaje(`Calificación guardada para ${alumno.account.name}`);
        setAlumnos(prev => prev.map(a => a.id === alumno.id ? { ...a, currentGrade: gradeValue } : a));
      } else {
        mostrarError('Error al guardar la calificación');
      }
    } catch (error) {
      console.error(error);
      mostrarError('Error al guardar la calificación');
    }
  };

  const guardarTodasCalificaciones = async () => {
    const calificacionesParaGuardar = alumnos.map(a => ({
      studentId: a.id,
      subjectId: materiaSeleccionada as number,
      grade: parseFloat(calificaciones[a.id]),
      notes: `Calificación masiva - ${new Date().toLocaleDateString()}`
    })).filter(c => !isNaN(c.grade) && c.grade >= 0 && c.grade <= 10);

    if (calificacionesParaGuardar.length === 0) {
      mostrarError('No hay calificaciones válidas para guardar');
      return;
    }

    try {
      const success = await guardarCalificacionesMasivo(calificacionesParaGuardar);
      if (success) {
        mostrarMensaje(`${calificacionesParaGuardar.length} calificaciones guardadas`);
        setAlumnos(prev => prev.map(a => {
          const cal = calificacionesParaGuardar.find(c => c.studentId === a.id);
          return cal ? { ...a, currentGrade: cal.grade } : a;
        }));
      } else {
        mostrarError('Error al guardar algunas calificaciones');
      }
    } catch (error) {
      console.error(error);
      mostrarError('Error al guardar calificaciones');
    }
  };

  const getEstadoTexto = (cal: string) => {
    const grade = parseFloat(cal);
    if (isNaN(grade)) return 'Sin calificar';
    if (grade >= 9) return 'Excelente';
    if (grade >= 7) return 'Bueno';
    if (grade >= 6) return 'Suficiente';
    return 'Reprobado';
  };

  const getEstadoColor = (cal: string) => {
    const grade = parseFloat(cal);
    if (isNaN(grade)) return 'default';
    if (grade >= 9) return 'success';
    if (grade >= 7) return 'info';
    if (grade >= 6) return 'warning';
    return 'error';
  };

  const mostrarMensaje = (msg: string) => setSnackbar({ open: true, message: msg, severity: 'success' });
  const mostrarError = (msg: string) => setSnackbar({ open: true, message: msg, severity: 'error' });
  const handleCloseSnackbar = () => setSnackbar(prev => ({ ...prev, open: false }));

  const calcularPromedio = (alumno: AlumnoParaCalificar) => {
    const grade = parseFloat(calificaciones[alumno.id]);
    return isNaN(grade) ? 0 : grade;
  };

  return (
    <Container maxWidth="xl">
      <Typography variant="h4" gutterBottom sx={{ mb: 3 }}>Evaluaciones - Secretaría</Typography>

      <Paper sx={{ p: 3, mb: 3 }}>
        <Typography variant="h6" gutterBottom sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
          <FilterIcon /> Filtros
        </Typography>

        {loadingFiltros && <Box sx={{ display: 'flex', justifyContent: 'center', my: 2 }}><CircularProgress /></Box>}

        <Grid container spacing={3}>
          <Grid item xs={12} md={3}>
            <FormControl fullWidth disabled={loadingFiltros}>
              <InputLabel>Carrera</InputLabel>
              <Select value={carreraSeleccionada} onChange={e => setCarreraSeleccionada(e.target.value as number)} label="Carrera">
                <MenuItem value="">Seleccionar carrera</MenuItem>
                {carreras.map(c => <MenuItem key={c.ID} value={c.ID}>{c.Name}</MenuItem>)}
              </Select>
            </FormControl>
          </Grid>

          <Grid item xs={12} md={3}>
            <FormControl fullWidth disabled={!carreraSeleccionada || loadingFiltros}>
              <InputLabel>Grupo</InputLabel>
              <Select value={grupoSeleccionado} onChange={e => setGrupoSeleccionado(e.target.value as string)} label="Grupo">
                <MenuItem value="">Seleccionar grupo</MenuItem>
                {grupos.map(g => <MenuItem key={g.id} value={g.id}>{g.nombre}</MenuItem>)}
              </Select>
            </FormControl>
          </Grid>

          <Grid item xs={12} md={3}>
            <FormControl fullWidth disabled={!grupoSeleccionado || loadingFiltros}>
              <InputLabel>Materia</InputLabel>
              <Select value={materiaSeleccionada} onChange={e => setMateriaSeleccionada(e.target.value as number)} label="Materia">
                <MenuItem value="">Seleccionar materia</MenuItem>
                {materias.map(m => <MenuItem key={m.ID} value={m.ID}>{m.Name}</MenuItem>)}
              </Select>
            </FormControl>
          </Grid>

          <Grid item xs={12} md={3}>
            <Button fullWidth variant="contained" startIcon={<SaveAllIcon />} onClick={guardarTodasCalificaciones} disabled={!materiaSeleccionada || alumnos.length === 0 || loading}>
              Guardar Todas
            </Button>
          </Grid>
        </Grid>
      </Paper>

      {alumnos.length > 0 && (
        <Paper sx={{ p: 3 }}>
          <Typography variant="h6" gutterBottom>Alumnos para Calificar ({alumnos.length})</Typography>
          {loading ? <Box sx={{ display: 'flex', justifyContent: 'center', p: 3 }}><CircularProgress /></Box> :
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>Matrícula</TableCell>
                  <TableCell>Nombre</TableCell>
                  <TableCell>Semestre</TableCell>
                  <TableCell>Calificación (0-10)</TableCell>
                  <TableCell>Estado</TableCell>
                  <TableCell>Promedio</TableCell>
                  <TableCell>Acciones</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {alumnos.map(a => (
                  <TableRow key={a.id}>
                    <TableCell>{a.student_id}</TableCell>
                    <TableCell>{a.account.name}</TableCell>
                    <TableCell><Chip label={`${a.semester}°`} size="small" color="primary" variant="outlined" /></TableCell>
                    <TableCell>
                      <TextField type="number" value={calificaciones[a.id] || ''} onChange={e => handleCalificacionChange(a.id, e.target.value)}
                        inputProps={{ min: 0, max: 10, step: 0.1, style: { textAlign: 'center' } }}
                        size="small" sx={{ width: 100 }} />
                    </TableCell>
                    <TableCell>
                      <Chip label={getEstadoTexto(calificaciones[a.id])} color={getEstadoColor(calificaciones[a.id])} size="small" />
                    </TableCell>
                    <TableCell>{calcularPromedio(a).toFixed(1)}</TableCell>
                    <TableCell>
                      <Tooltip title="Guardar calificación">
                        <Button variant="outlined" size="small" startIcon={<SaveIcon />} onClick={() => guardarCalificacionIndividual(a)} disabled={!calificaciones[a.id]}>
                          Guardar
                        </Button>
                      </Tooltip>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>}
        </Paper>
      )}

      {!loading && alumnos.length === 0 && carreraSeleccionada && grupoSeleccionado && materiaSeleccionada && (
        <Paper sx={{ p: 3, textAlign: 'center' }}>
          <Typography>No hay alumnos inscritos en esta materia con los filtros seleccionados</Typography>
        </Paper>
      )}

      <Snackbar open={snackbar.open} autoHideDuration={4000} onClose={handleCloseSnackbar} anchorOrigin={{ vertical: 'bottom', horizontal: 'center' }}>
        <Alert onClose={handleCloseSnackbar} severity={snackbar.severity} sx={{ width: '100%' }}>
          {snackbar.message}
        </Alert>
      </Snackbar>
    </Container>
  );
};

export default Evaluaciones;
