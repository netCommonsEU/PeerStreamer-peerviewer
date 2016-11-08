import React, {PropTypes as T} from 'react';
import Paper from 'material-ui/Paper';

const styles = {
    title: {
        fontSize: '1.3em',
        padding: '.3em 0'
    },
    paper: {
        padding: '2em'
    }
}

const StreamInfo = ({title}) => {
    return (
        <Paper style={styles.paper}>
            <div style={styles.title}>{title}</div>
        </Paper>
    );
};

export default StreamInfo;